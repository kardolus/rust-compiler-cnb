package rust

import (
	"os"
	"path/filepath"

	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/cloudfoundry/libcfbuildpack/layers"
)

type Runner interface {
	Run(bin, dir string, args ...string) error
}

type Logger interface {
	Info(format string, args ...interface{})
}

type Rust struct {
	Runner Runner
	Logger Logger
}

func (r Rust) Install(location string, layer layers.DependencyLayer, artifact string) error {
	layer.Logger.SubsequentLine("Expanding to %s", layer.Root)
	if err := helper.ExtractTarGz(artifact, layer.Root, 1); err != nil {
		return err
	}

	// TODO grep version from buildpack.yml
	layer.Logger.SubsequentLine("Installing Rust Components")
	if err := r.Runner.Run("sh", layer.Root, "rustup-init.sh", "-y", "--default-toolchain", "nightly"); err != nil {
		return err
	}

	cargoBin := filepath.Join(os.Getenv("HOME"), ".cargo", "bin", "cargo")

	layer.Logger.SubsequentLine("Building app from %s", location)
	if err := r.Runner.Run(cargoBin, location, "build"); err != nil {
		return err
	}

	return nil
}
