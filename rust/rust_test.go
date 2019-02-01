package rust_test

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/libcfbuildpack/helper"
	"github.com/cloudfoundry/libcfbuildpack/layers"
	"github.com/cloudfoundry/libcfbuildpack/test"
	"github.com/kardolus/rust-cnb/rust"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

//go:generate mockgen -source=rust.go -destination=mock_test.go -package=rust_test

func TestUnitRust(t *testing.T) {
	spec.Run(t, "Rust", testRust, spec.Report(report.Terminal{}))
}

func testRust(t *testing.T, when spec.G, it spec.S) {
	var (
		mockCtrl   *gomock.Controller
		mockRunner *MockRunner
		mockLogger *MockLogger
		pkgManager rust.Rust
		factory    *test.DetectFactory
		layer      layers.DependencyLayer
	)

	it.Before(func() {
		RegisterTestingT(t)
		factory = test.NewDetectFactory(t)
		mockCtrl = gomock.NewController(t)
		mockRunner = NewMockRunner(mockCtrl)
		mockLogger = NewMockLogger(mockCtrl)

		pkgManager = rust.Rust{Runner: mockRunner, Logger: mockLogger}
	})

	it.After(func() {
		mockCtrl.Finish()
	})

	when("installing and buildpack.yml exists", func() {
		it("grabs the rust version from buildpack.yml if it is present", func() {
			const version string = "1.2.3"
			buildpackYAMLString := fmt.Sprintf("rust:\n  version: %s", version)
			Expect(helper.WriteFile(filepath.Join(factory.Detect.Application.Root, "buildpack.yml"), 0666, buildpackYAMLString)).To(Succeed())

			mockRunner.EXPECT().Run("sh", layer.Root, "rustup-init.sh", "-y", "--default-toolchain", version)
			mockRunner.EXPECT().Run(rust.CargoBin, factory.Detect.Application.Root, "build")

			Expect(pkgManager.Install(factory.Detect.Application.Root, layer)).To(Succeed())
		})

		it("grabs the default rust version if it is not present", func() {
			mockRunner.EXPECT().Run("sh", layer.Root, "rustup-init.sh", "-y")
			mockRunner.EXPECT().Run(rust.CargoBin, factory.Detect.Application.Root, "build")

			Expect(pkgManager.Install(factory.Detect.Application.Root, layer)).To(Succeed())
		})
	})
}
