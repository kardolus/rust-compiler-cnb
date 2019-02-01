package rust

import (
	"os"
	"path/filepath"

	"github.com/cloudfoundry/libcfbuildpack/layers"
	"github.com/kardolus/rust-cnb/utils"
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

var CargoBin = filepath.Join(os.Getenv("HOME"), ".cargo", "bin", "cargo")

func (r Rust) Install(location string, layer layers.DependencyLayer) error {
	config, err := utils.ParseConfig(location)

	if err != nil {
		return err
	}
	version := config.Rust.Version

	layer.Logger.SubsequentLine("Installing Rust Components")
	if version != "" {
		if err := r.Runner.Run("sh", layer.Root, "rustup-init.sh", "-y", "--default-toolchain", version); err != nil {
			return err
		}
	} else {
		if err := r.Runner.Run("sh", layer.Root, "rustup-init.sh", "-y"); err != nil {
			return err
		}
	}

	layer.Logger.SubsequentLine("Building app from %s", location)
	if err := r.Runner.Run(CargoBin, location, "build"); err != nil {
		return err
	}

	return nil
}
