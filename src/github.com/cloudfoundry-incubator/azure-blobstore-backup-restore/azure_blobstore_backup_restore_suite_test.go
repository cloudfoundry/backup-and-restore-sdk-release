package azure_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestAzureBlobstoreBackupRestore(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "AzureBlobstoreBackupRestore Suite")
}
