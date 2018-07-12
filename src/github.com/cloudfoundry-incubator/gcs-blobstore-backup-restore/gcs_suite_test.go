package gcs_test

import (
	"testing"

	"time"

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
	serviceAccountKeyJson = MustHaveEnv("GCP_SERVICE_ACCOUNT_KEY")
	Authenticate(serviceAccountKeyJson)
})
