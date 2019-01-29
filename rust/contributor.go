package rust

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/buildpack/libbuildpack/application"
	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/cloudfoundry/libcfbuildpack/layers"
)

const (
	Dependency = "rustup"
	Cache      = "cache"
)

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
	Install(location string, layer layers.DependencyLayer, arifact string) error
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

	lockFile := filepath.Join(context.Application.Root, "Cargo.lock")
	if exists, err := helper.FileExists(lockFile); err != nil {
		return Contributor{}, false, err
	} else if !exists {
		return Contributor{}, false, fmt.Errorf(`unable to find "Cargo.lock"`)
	}

	buf, err := ioutil.ReadFile(lockFile)
	if err != nil {
		return Contributor{}, false, err
	}

	hash := sha256.Sum256(buf)

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
	// TODO this should use a downloadlayer instead
	return c.packagesLayer.Contribute(func(artifact string, layer layers.DependencyLayer) error {

		if err := c.manager.Install(c.packagesLayer.Root, layer, artifact); err != nil {
			return err
		}

		return c.launchLayer.WriteMetadata(layers.Metadata{Processes: []layers.Process{{"web", "cargo run"}}})
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
