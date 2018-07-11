package gcs_test

import (
	"testing"

	"time"

	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var serviceAccountKeyJson string

func TestGcsBlobstoreBackupRestore(t *testing.T) {
	RegisterFailHandler(Fail)
	SetDefaultEventuallyTimeout(15 * time.Minute)
	RunSpecs(t, "GCS Suite")
}

var _ = BeforeSuite(func() {
	Authenticate(MustHaveEnv("GCP_SERVICE_ACCOUNT_KEY_PATH"))

	serviceAccountKeyPath := MustHaveEnv("GCP_SERVICE_ACCOUNT_KEY_PATH")
	serviceAccountKeyJsonBytes, err := ioutil.ReadFile(serviceAccountKeyPath)
	Expect(err).NotTo(HaveOccurred())
	serviceAccountKeyJson = string(serviceAccountKeyJsonBytes)
})
