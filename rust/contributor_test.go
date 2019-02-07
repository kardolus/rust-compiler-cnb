package rust_test

import (
	"path/filepath"
	"testing"

	"github.com/buildpack/libbuildpack/buildplan"
	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/cloudfoundry/libcfbuildpack/test"
	"github.com/golang/mock/gomock"
	"github.com/kardolus/rust-cnb/rust"
	"github.com/kardolus/rust-cnb/utils"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

//go:generate mockgen -source=rust.go -destination=mock_test.go -package=rust_test

func TestUnitContributor(t *testing.T) {
	spec.Run(t, "Contributor", testContributor, spec.Report(report.Terminal{}))
}

func testContributor(t *testing.T, when spec.G, it spec.S) {
	var (
		stubRustFixture string
		factory         *test.BuildFactory
		pkgManager      rust.Rust
		mockCtrl        *gomock.Controller
		mockRunner      *MockRunner
		mockLogger      *MockLogger
	)

	it.Before(func() {
		RegisterTestingT(t)
		factory = test.NewBuildFactory(t)
		mockCtrl = gomock.NewController(t)
		mockRunner = NewMockRunner(mockCtrl)
		mockLogger = NewMockLogger(mockCtrl)

		stubRustFixture = filepath.Join("testdata", "stub-rust.tar.gz")
		pkgManager = rust.Rust{Runner: mockRunner, Logger: mockLogger}

		Expect(helper.WriteFile(filepath.Join(factory.Build.Application.Root, utils.CARGO_TOML), 0666, "")).To(Succeed())
	})

	it.After(func() {
		mockCtrl.Finish()
	})

	it("returns true if build plan exists and version is set", func() {
		factory.AddDependency(rust.Dependency, stubRustFixture)
		factory.AddBuildPlan(rust.Dependency, buildplan.Dependency{
			Version: "*",
		})

		_, ok, err := rust.NewContributor(factory.Build, pkgManager)
		Expect(ok).To(BeTrue())
		Expect(err).NotTo(HaveOccurred())
	})

	it("returns false if build plan does not exist", func() {
		_, ok, err := rust.NewContributor(factory.Build, pkgManager)
		Expect(ok).To(BeFalse())
		Expect(err).NotTo(HaveOccurred())
	})

	it("returns false if build plan exists but version is not set", func() {
		factory.AddDependency(rust.Dependency, stubRustFixture)
		factory.AddBuildPlan(rust.Dependency, buildplan.Dependency{})

		_, ok, err := rust.NewContributor(factory.Build, pkgManager)
		Expect(ok).To(BeFalse())
		Expect(err).NotTo(HaveOccurred())
	})

	it("contributes Rustup to build", func() {
		factory.AddDependency(rust.Dependency, stubRustFixture)
		factory.AddBuildPlan(rust.Dependency, buildplan.Dependency{
			Metadata: buildplan.Metadata{"build": true},
			Version:  "*",
		})

		c, shouldContribute, err := rust.NewContributor(factory.Build, pkgManager)
		Expect(err).NotTo(HaveOccurred())
		Expect(shouldContribute).To(BeTrue())

		mockRunner.EXPECT().Run("sh", filepath.Join(factory.Build.Layers.Root, "rustup"), "rustup-init.sh", "-y")
		mockRunner.EXPECT().Run(rust.CargoBin, factory.Build.Application.Root, "build")

		Expect(c.Contribute()).To(Succeed())

		layer := factory.Build.Layers.Layer("rustup")
		Expect(layer).To(test.HaveLayerMetadata(true, true, false))
		Expect(filepath.Join(layer.Root, "stub.txt")).To(BeARegularFile())
	})

	it("uses Cargo.lock for identity", func() {
		test.WriteFile(t, filepath.Join(factory.Build.Application.Root, "Cargo.lock"), "cargo lock")
		factory.AddDependency(rust.Dependency, stubRustFixture)
		factory.AddBuildPlan(rust.Dependency, buildplan.Dependency{
			Version: "*",
		})

		contributor, ok, err := rust.NewContributor(factory.Build, pkgManager)
		Expect(ok).To(BeTrue())
		Expect(err).NotTo(HaveOccurred())

		name, version := contributor.CacheMetadata.Identity()
		Expect(name).To(Equal(rust.Dependency))
		Expect(version).To(Equal("91c4d0a3ab83742413103cf2ba6a38803f8d8c598f8e1299bb9e31969feb6dd6"))
	})
}
