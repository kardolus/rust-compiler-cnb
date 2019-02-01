# Rust Cloud Native Buildpack
To package this buildpack for consumption:
```
$ ./scripts/package.sh
```
This builds the buildpack's Go source using GOOS=linux by default. You can supply another value as the first argument to package.sh.

## Quick test
- Pull the stacks
```
docker pull cfbuildpacks/cflinuxfs3-cnb-experimental:build
docker pull cfbuildpacks/cflinuxfs3-cnb-experimental:run
```
- Pull the cflinuxfs3 builder: `docker pull kardolus/fs3builder`
- Package the buildpack: `./scripts/package.sh`
- Get the pack-cli: `./scripts/install_tools.sh`
- Create an OCI image:
```
pack build rustapp --builder kardolus/fs3builder --buildpack </path/to/packaged/buildpack> -p integration/testdata/simple_app/ --no-pull
```
- Run the image: `docker run -it rustapp`
