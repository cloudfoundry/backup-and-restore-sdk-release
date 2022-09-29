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
	RunSpecs(t, "S3 Main Suite")
}

var _ = SynchronizedBeforeSuite(func() []byte {
	binaryPath, err := gexec.Build("s3-blobstore-backup-restore/cmd/s3-blobstore-backup-restore", "-mod", "vendor", "-buildvcs=false")
	Expect(err).NotTo(HaveOccurred())

	return []byte(binaryPath)

}, func(rawBinaryPath []byte) {
	binaryPath = string(rawBinaryPath)
})
