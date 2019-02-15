package rust

import (
	"os"
	"path/filepath"

	"github.com/cloudfoundry/libcfbuildpack/helper"
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

var (
	CargoHome = filepath.Join(os.Getenv("HOME"), ".cargo")
	CargoBin  = filepath.Join(CargoHome, "bin", "cargo")
)

func (r Rust) Install(location string, layer layers.Layer, cacheLayer layers.Layer) error {
	if err := r.moveDir(cacheLayer.Root, location, TargetDir); err != nil {
		return err
	}
	if err := r.moveDir(cacheLayer.Root, CargoHome, RegistryDir); err != nil {
		return err
	}

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
	if err := r.Runner.Run(CargoBin, location, "build", "--release"); err != nil {
		return err
	}

	return nil
}

func (r Rust) moveDir(source, target, name string) error {
	dir := filepath.Join(source, name)
	dest := filepath.Join(target, name)

	if exists, err := helper.FileExists(dir); err != nil {
		return err
	} else if !exists {
		return nil
	}

	r.Logger.Info("Reusing existing %s directory", name)
	if err := helper.CopyDirectory(dir, dest); err != nil {
		return err
	}

	return nil
}
