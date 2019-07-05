package contract_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestAzureBlobstoreBackupRestore(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "AzureBlobstoreBackupRestore Suite")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	MustHaveEnv("AZURE_STORAGE_ACCOUNT")
	MustHaveEnv("AZURE_STORAGE_KEY")
	MustHaveEnv("AZURE_STORAGE_ACCOUNT_NO_SOFT_DELETE")
	MustHaveEnv("AZURE_STORAGE_KEY_NO_SOFT_DELETE")
	return nil
}, func([]byte) {})

func MustHaveEnv(keyname string) string {
	val := os.Getenv(keyname)
	Expect(val).NotTo(BeEmpty(), "Need "+keyname+" for the test")
	return val
}
