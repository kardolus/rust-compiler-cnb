package integration

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
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

			imageName := "rust_test_image"
			fixture := filepath.Join("testdata", "simple_app")

			cmd := exec.Command("pack", "build", imageName, "--builder", "kardolus/fs3builder", "--buildpack", bp, "-p", fixture, "--no-pull", "--clear-cache")
			cmd.Stderr = os.Stderr
			cmd.Stdout = os.Stdout
			Expect(cmd.Run()).To(Succeed())

			defer cleanUp(imageName)

			cmd = exec.Command("docker", "run", "-p", "8080:8080", imageName)
			cmd.Stderr = os.Stderr
			cmd.Stdout = os.Stdout
			Expect(cmd.Run()).To(Succeed())

			Expect(HTTPGet()).To(Succeed())
		})
	})
}

func cleanUp(imageName string) {
	cmd := exec.Command("docker", "rmi", imageName, "-f")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	Expect(cmd.Run()).To(Succeed())
}

func HTTPGet() error {
	resp, err := http.Get("http://localhost:8080")
	if err != nil {
		return err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("received bad response from application")
	}

	return nil
}
