# LXB - LXD Image Builder

LXB is an attempt to bring a little of the Docker image creation magic to LXD with an automated, templated build file.

Use `go get` to install: `go get github.com/colebrumley/lxb`. Alternatively, use `glide` to make sure your dependencies are up to date with mine:
```bash
go get github.com/Masterminds/glide
go get -d github.com/colebrumley/lxb
cd $GOPATH/src/github.com/colebrumley/lxb
export GO15VENDOREXPERIMENT=1
glide up
go build
```

Currently you must run LXB on a host running LXD since LXB is hard coded to look for a local unix socket. In the future I'd like to add support for calling remote daemons. You can however _work_ on a remote via the local daemon with the `--remote` flag.

## usage
```bash
NAME:
   lxb - LXD Image Builder

USAGE:
   lxb [global options] [arguments...]

VERSION:
   0.0.2

GLOBAL OPTIONS:
   --lxfile, -f "lxfile.yml"  Path to the build spec
   --context, -c "./"   Path to the build context
   --keep, -k     Don't remove the build container when complete
   --verbose      Print extra debugging output
   --remote "local"   LXD daemon address [$LXB_REMOTE]
   --version, -v    print the version
```
The process is similar to building a Docker image. Call `lxb` either in the build context directory (only important if your `lxfile.yml` uses `files`), or pass the `-c` flag.

The build config can be supplied by providing nothing (in which case `./lxfile.yml` is loaded), the `--lxfile` flag, or by passing `-` as the first argument (which will read from stdin).

## lxfile usage
Lxfiles are YAML files (just for you @jimmymac) that contain specifications for both the build container and the resulting image.

The only key that is _strictly_ required is `baseimg`, all other keys will be ignored if they are omitted.

**Example:**

```yaml
baseimg: trusty
image_properties:
  description: Apache2 on Ubuntu Trusty x64
image_aliases:
  - trusty-apache2
public: true
build_profiles:
  - default
build_config:
  # See https://github.com/lxc/lxd/blob/master/specs/configuration.md#container-configuration
  limits.memory: 512MB
env:
  BUILD_ENV: dev
  CONTAINERIZED: true
  TERM: xterm-color
cmd:
  # A list of commands that will be run inside the container (using /bin/sh)
  - apt-get update
  - apt-get -y install apache2
  - apt-get -y clean
```
