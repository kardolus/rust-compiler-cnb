package integration

import (
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/dagger"

	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	. "github.com/onsi/gomega"
)

func TestIntegration(t *testing.T) {
	spec.Run(t, "Integration", testIntegration, spec.Report(report.Terminal{}))
}

func testIntegration(t *testing.T, when spec.G, it spec.S) {
	it.Before(func() {
		RegisterTestingT(t)
	})

	when("when a simple app has no caching", func() {
		it("should build a working OCI image", func() {
			bp, err := dagger.PackageBuildpack()
			Expect(err).ToNot(HaveOccurred())

			app, err := dagger.PackBuild(filepath.Join("testdata", "simple_app"), bp)
			Expect(err).ToNot(HaveOccurred())
			defer app.Destroy()

			Expect(app.Start()).To(Succeed())
			// Expect(app.HTTPGet("/")).To(Succeed())
		})
	})
}
