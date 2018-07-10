package gcs_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestGcsBlobstoreBackupRestore(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GCS Suite")
}
