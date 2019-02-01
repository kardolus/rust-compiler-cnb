package utils

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/BurntSushi/toml"

	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/go-yaml/yaml"
)

const (
	CARGO_TOML = "Cargo.toml"
	CARGO_LOCK = "Cargo.lock"
)

type CommandRunner struct {
}

func (r CommandRunner) Run(bin, dir string, args ...string) error {
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

type Config struct {
	Rustup struct {
		Version string `yaml:"version"`
	} `yaml:"rustup"`
	Rust struct {
		Version string `yaml:"version"`
	} `yaml:"rust"`
}

type AppMetadata struct {
	Package PackageMetadata `toml:"package"`
}

type PackageMetadata struct {
	Name    string
	Version string
	Authors []string
}

func ParseConfig(appDir string) (Config, error) {
	config := Config{}
	buildpackYAMLPath := filepath.Join(appDir, "buildpack.yml")
	fmt.Println("Path: ", buildpackYAMLPath)
	exists, err := helper.FileExists(buildpackYAMLPath)
	if err != nil {
		return config, err
	}

	if exists {
		buf, err := ioutil.ReadFile(buildpackYAMLPath)
		if err != nil {
			return config, err
		}

		if err := yaml.Unmarshal(buf, &config); err != nil {
			return config, err
		}
	}
	return config, nil
}

func ParseAppMetadata(appDir string) (AppMetadata, error) {
	var meta AppMetadata
	buf, err := ioutil.ReadFile(filepath.Join(appDir, CARGO_TOML))
	if err != nil {
		return meta, err
	}

	tomlData := string(buf)
	if _, err := toml.Decode(tomlData, &meta); err != nil {
		return meta, err
	}

	return meta, nil
}
