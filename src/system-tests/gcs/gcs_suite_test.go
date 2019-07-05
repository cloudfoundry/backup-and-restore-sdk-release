package gcs_test

import (
	"testing"

	"time"

	"os/exec"

	. "system-tests"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	. "github.com/onsi/gomega/gexec"
)

func TestGcs(t *testing.T) {
	RegisterFailHandler(Fail)
	SetDefaultEventuallyTimeout(15 * time.Minute)
	RunSpecs(t, "GCS System Tests Suite")
}

var _ = BeforeSuite(func() {
	MustRunSuccessfully("gcloud", "auth", "activate-service-account",
		"--key-file", MustHaveEnv("GCP_SERVICE_ACCOUNT_KEY_PATH"))
	MustRunSuccessfully("gcloud", "config", "set", "project", MustHaveEnv("GCP_PROJECT_NAME"))
})

func MustRunSuccessfully(command string, args ...string) {
	cmd := exec.Command(command, args...)

	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(Exit(0))
}

func Run(command string, args ...string) *gexec.Session {
	cmd := exec.Command(command, args...)

	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	return session
}
