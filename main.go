package main

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/codegangsta/cli"
	"github.com/lxc/lxd"
	log "github.com/sirupsen/logrus"
)

// AppVersion is the global application version
var AppVersion string

func main() {
	app := cli.NewApp()
	app.Name = "lxb"
	app.HelpName = "lxb"
	app.Version = AppVersion
	app.HideHelp = true
	app.Usage = "LXD Image Builder"
	app.ArgsUsage = ""
	app.Action = build
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "lxfile,f",
			Usage: "Path to the build spec",
			Value: "lxfile.yml",
		},
		cli.StringFlag{
			Name:  "context,c",
			Usage: "Path to the build context",
			Value: "./",
		},
		cli.BoolFlag{
			Name:  "keep,k",
			Usage: "Don't remove the build container when complete",
		},
		cli.BoolFlag{
			Name:  "verbose",
			Usage: "Print extra debugging output",
		},
		cli.StringFlag{
			Name:   "remote",
			Usage:  "LXD daemon address",
			Value:  "local",
			EnvVar: "LXB_REMOTE",
		},
	}
	app.Run(os.Args)
}

func build(c *cli.Context) {
	var (
		lxfile       = c.String("lxfile")
		buildContext = c.String("context")
		remote       = c.GlobalString("remote")
	)
	// Validate args
	if c.GlobalBool("verbose") {
		log.SetLevel(log.DebugLevel)
	}

	// validate the context dir
	contextAbsPath := asrt(filepath.Abs(buildContext)).(string)
	if !dirExists(contextAbsPath) {
		log.Errorln("Build context directory does not exist!")
	}

	// move to context dir
	if err := os.Chdir(contextAbsPath); err != nil {
		log.Error(err)
	}

	// Load the build spec
	if c.Args().First() == "-" {
		b, err := ioutil.ReadAll(os.Stdin)
		lxfile = string(b)
		if err != nil {
			log.Error(err)
		}
	} else if !fileExists(lxfile) {
		log.Errorln("No lxfile found!")
		os.Exit(1)
	}
	spec := LoadBuildSpec(lxfile)
	log.Debugln("Loaded build spec")

	// connect to LXD
	configDir := "$HOME/.config/lxc"
	if os.Getenv("LXD_CONF") != "" {
		configDir = os.Getenv("LXD_CONF")
	}
	configPath := os.ExpandEnv(path.Join(configDir, "config.yml"))
	config, err := lxd.LoadConfig(configPath)
	if err != nil {
		log.Debug(err)
		config = &lxd.DefaultConfig
	}

	realRemote := config.ParseRemote(remote)
	log.Debugf("Connecting to remote %s", realRemote)
	cl, err := lxd.NewClient(config, realRemote)
	if err != nil {
		log.Error(err)
	}
	if !cl.AmTrusted() {
		log.Errorln("Connection is not trusted by LXD! Check your cert and key.")
	}
	log.Debugln("Connected to LXD")

	// create build object
	b := NewBuild(spec, cl, remote)
	log.Infoln("Starting build")
	if err := b.Execute(c.Bool("keep")); err != nil {
		log.Error(err)
	}

	log.Println("Build complete!")
}
