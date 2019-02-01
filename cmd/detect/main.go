package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/libcfbuildpack/detect"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/kardolus/rust-cnb/rust"
	"github.com/kardolus/rust-cnb/utils"
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

	exists, err := helper.FileExists(filepath.Join(context.Application.Root, utils.CARGO_TOML))
	if err != nil {
		return detect.FailStatusCode, err
	} else if !exists {
		return detect.FailStatusCode, errors.New("no Cargo.toml found!")
	}

	config, err := utils.ParseConfig(context.Application.Root)
	if err != nil {
		return detect.FailStatusCode, err
	}
	version = config.Rustup.Version

	return context.Pass(buildplan.BuildPlan{
		rust.Dependency: buildplan.Dependency{
			Version:  version,
			Metadata: buildplan.Metadata{"build": true, "launch": true},
		},
	})
}
