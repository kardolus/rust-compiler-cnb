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

	if err := r.Runner.Run("sh", layer.Root, "rustup-init.sh", "-y"); err != nil {
		return err
	}

	// TODO grep version from buildpack.yml
	// TODO can the rust version be set on the first invocation?
	// TODO do not hardcode home directory
	// Current error: Cargo needs gcc
	homeDir := "/home/pack"
	installDir := filepath.Join(layer.Root, "install")
	cargoDir := filepath.Join(installDir, "cargo")
	multirustDir := filepath.Join(installDir, "multirust")

	os.MkdirAll(installDir, 0777)
	os.MkdirAll(cargoDir, 0777)
	os.MkdirAll(multirustDir, 0777)

	if err := layer.OverrideSharedEnv("RUSTUP_HOME", multirustDir); err != nil {
		return err
	}

	if err := layer.OverrideSharedEnv("CARGO_HOME", cargoDir); err != nil {
		return err
	}

	if err := layer.AppendSharedEnv("PATH", filepath.Join(cargoDir, "bin")+":"); err != nil {
		return err
	}

	if err := r.Runner.Run(".cargo/bin/rustup", homeDir, "default", "nightly"); err != nil {
		return err
	}

	if err := helper.CopyDirectory(filepath.Join(homeDir, ".cargo"), cargoDir); err != nil {
		return err
	}

	if err := helper.CopyDirectory(filepath.Join(homeDir, ".multirust"), multirustDir); err != nil {
		return err
	}

	return nil
}
