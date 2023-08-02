package contract_test

import (
	"testing"

	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGcsBlobstoreBackupRestore(t *testing.T) {
	RegisterFailHandler(Fail)
	SetDefaultEventuallyTimeout(15 * time.Minute)
	RunSpecs(t, "GCS Contract Tests")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	Authenticate(MustHaveEnv("GCP_SERVICE_ACCOUNT_KEY"), MustHaveEnv("GCP_PROJECT_NAME"))
	return nil
}, func([]byte) {})
