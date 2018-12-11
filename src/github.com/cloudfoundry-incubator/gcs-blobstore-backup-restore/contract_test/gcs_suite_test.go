package contract_test

import (
	"testing"

	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestGcsBlobstoreBackupRestore(t *testing.T) {
	RegisterFailHandler(Fail)
	SetDefaultEventuallyTimeout(15 * time.Minute)
	RunSpecs(t, "GCS Suite")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	Authenticate(MustHaveEnv("GCP_SERVICE_ACCOUNT_KEY"))
	return nil
}, func([]byte) {})
