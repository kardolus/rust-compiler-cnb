package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/libcfbuildpack/detect"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/kardolus/rust-cnb/rust"
	"gopkg.in/yaml.v2"
)

func main() {
	context, err := detect.DefaultDetect()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to create default detect context: %s", err)
		os.Exit(100)
	}

	code, err := runDetect(context)
	if err != nil {
		context.Logger.Info(err.Error())
	}
	os.Exit(code)
}

func runDetect(context detect.Detect) (int, error) {
	var version string

	exists, err := helper.FileExists(filepath.Join(context.Application.Root, "Cargo.toml"))
	if err != nil {
		return detect.FailStatusCode, err
	} else if !exists {
		return detect.FailStatusCode, errors.New("no Cargo.toml found!")
	}

	buildpackYAMLPath := filepath.Join(context.Application.Root, "buildpack.yml")
	exists, err = helper.FileExists(buildpackYAMLPath)
	if err != nil {
		return detect.FailStatusCode, err
	}

	if exists {
		buf, err := ioutil.ReadFile(buildpackYAMLPath)
		if err != nil {
			return detect.FailStatusCode, err
		}

		config := struct {
			Rust struct {
				Version string `yaml:"version"`
			} `yaml:"rust"`
		}{}
		if err := yaml.Unmarshal(buf, &config); err != nil {
			return detect.FailStatusCode, err
		}

		version = config.Rust.Version
	}

	return context.Pass(buildplan.BuildPlan{
		rust.Dependency: buildplan.Dependency{
			Version:  version,
			Metadata: buildplan.Metadata{"build": true, "launch": true},
		},
	})
}
