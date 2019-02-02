package rust

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/kardolus/rust-cnb/utils"

	"github.com/buildpack/libbuildpack/application"
	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/cloudfoundry/libcfbuildpack/layers"
)

const (
	Dependency = "rustup"
	Cache      = "cache"
)

var CargoToml string

type Contributor struct {
	CacheMetadata      Metadata
	RustMetadata       Metadata
	manager            PackageManager
	app                application.Application
	packagesLayer      layers.DependencyLayer
	launchLayer        layers.Layers
	buildContribution  bool
	launchContribution bool
}

type PackageManager interface {
	Install(location string, layer layers.DependencyLayer) error
}

type Metadata struct {
	Name string
	Hash string
}

func (m Metadata) Identity() (name string, version string) {
	return m.Name, m.Hash
}

func NewContributor(context build.Build, manager PackageManager) (Contributor, bool, error) {
	plan, wantDependency := context.BuildPlan[Dependency]
	if !wantDependency {
		return Contributor{}, false, nil
	}

	deps, err := context.Buildpack.Dependencies()
	if err != nil {
		return Contributor{}, false, err
	}

	dep, err := deps.Best(Dependency, plan.Version, context.Stack)
	if err != nil {
		return Contributor{}, false, err
	}

	CargoToml = filepath.Join(context.Application.Root, utils.CARGO_TOML)
	if exists, err := helper.FileExists(CargoToml); err != nil {
		return Contributor{}, false, err
	} else if !exists {
		return Contributor{}, false, fmt.Errorf("unable to find " + utils.CARGO_TOML)
	}

	cargoLock := filepath.Join(context.Application.Root, utils.CARGO_LOCK)

	var hash [32]byte
	if _, err := os.Stat(cargoLock); err == nil {
		buf, err := ioutil.ReadFile(cargoLock)
		if err != nil {
			return Contributor{}, false, err
		}
		hash = sha256.Sum256(buf)
	}

	// TODO implement caching
	contributor := Contributor{
		manager:       manager,
		app:           context.Application,
		packagesLayer: context.Layers.DependencyLayer(dep),
		launchLayer:   context.Layers,
		CacheMetadata: Metadata{Cache, hex.EncodeToString(hash[:])},
		RustMetadata:  Metadata{"org.cloudfoundry.buildpacks.rust", hex.EncodeToString(hash[:])},
	}

	if _, ok := plan.Metadata["build"]; ok {
		contributor.buildContribution = true
	}

	if _, ok := plan.Metadata["launch"]; ok {
		contributor.launchContribution = true
	}
	return contributor, true, nil
}

func (c Contributor) Contribute() error {
	return c.packagesLayer.Contribute(func(artifact string, layer layers.DependencyLayer) error {
		layer.Logger.SubsequentLine("Expanding to %s", layer.Root)
		if err := helper.ExtractTarGz(artifact, layer.Root, 1); err != nil {
			return err
		}

		if err := c.manager.Install(c.app.Root, layer); err != nil {
			return err
		}

		meta, err := utils.ParseAppMetadata(c.app.Root)
		if err != nil {
			return err
		}

		return c.launchLayer.WriteMetadata(layers.Metadata{Processes: []layers.Process{{"web", filepath.Join(c.app.Root, "target", "release", meta.Package.Name)}}})
	}, c.flags()...)
}

func (c Contributor) flags() []layers.Flag {
	flags := []layers.Flag{layers.Cache}

	if c.buildContribution {
		flags = append(flags, layers.Build)
	}

	if c.launchContribution {
		flags = append(flags, layers.Launch)
	}
	return flags
}
