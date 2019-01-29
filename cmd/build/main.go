package main

import (
	"fmt"
	"os"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/libcfbuildpack/build"
	"github.com/kardolus/rust-cnb/rust"
	"github.com/kardolus/rust-cnb/utils"
)

func main() {
	context, err := build.DefaultBuild()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to create default build context: %s", err)
		os.Exit(100)
	}

	code, err := runBuild(context)
	if err != nil {
		context.Logger.Info(err.Error())
	}
	os.Exit(code)
}

func runBuild(context build.Build) (int, error) {
	context.Logger.FirstLine(context.Logger.PrettyIdentity(context.Buildpack))

	packageManager := rust.Rust{
		Runner: utils.CommandRunner{},
		Logger: context.Logger,
	}

	contributor, willContribute, err := rust.NewContributor(context, packageManager)
	if err != nil {
		return context.Failure(102), err
	}

	if willContribute {
		if err := contributor.Contribute(); err != nil {
			return context.Failure(103), err
		}
	}

	return context.Success(buildplan.BuildPlan{})
}
