package contract_test

import (
	"os"
	"testing"

	"azure-blobstore-backup-restore"

	. "github.com/onsi/ginkgo/v2"
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

	// The specs only create containers in these two accounts.
	endpointSuffix, _ := azure.DefaultEnvironment.Suffix()
	deleteLeftoverContainers(MustHaveEnv("AZURE_STORAGE_ACCOUNT"), MustHaveEnv("AZURE_STORAGE_KEY"), endpointSuffix)
	if account := os.Getenv("AZURE_DIFFERENT_STORAGE_ACCOUNT"); account != "" {
		deleteLeftoverContainers(account, os.Getenv("AZURE_DIFFERENT_STORAGE_KEY"), endpointSuffix)
	}
	return nil
}, func([]byte) {})

func MustHaveEnv(keyname string) string {
	val := os.Getenv(keyname)
	Expect(val).NotTo(BeEmpty(), "Need "+keyname+" for the test")
	return val
}
