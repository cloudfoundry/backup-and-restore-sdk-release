package main_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var binaryPath string

func TestGcsBlobstoreBackupRestore(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GCS Main Suite")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	binaryPath, err := gexec.Build("gcs-blobstore-backup-restore/cmd/gcs-blobstore-backup-restore")
	Expect(err).NotTo(HaveOccurred())

	return []byte(binaryPath)

}, func(rawBinaryPath []byte) {
	binaryPath = string(rawBinaryPath)
})
