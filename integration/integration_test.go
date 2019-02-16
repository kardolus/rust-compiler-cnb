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
	var bp string

	it.Before(func() {
		RegisterTestingT(t)

		var err error
		bp, err = dagger.PackageBuildpack()
		Expect(err).ToNot(HaveOccurred())
	})

	when("when ran with a simple app", func() {
		it("should build a working OCI image", func() {
			app, err := dagger.PackBuild(filepath.Join("testdata", "simple_app"), bp)
			Expect(err).ToNot(HaveOccurred())
			defer app.Destroy()

			Expect(app.Start()).To(Succeed())

			_, _, err = app.HTTPGet("/")
			Expect(err).NotTo(HaveOccurred())
		})
	})
}
