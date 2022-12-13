package incremental_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"s3-blobstore-backup-restore/incremental"
)

var _ = Describe("BackedUpBlob", func() {
	It("knows the live blob path", func() {
		blob := incremental.BackedUpBlob{
			Path:                "timestamp/bucket_id/fd/f0/blob1/uuid",
			BackupDirectoryPath: "timestamp/bucket_id",
		}

		path := blob.LiveBlobPath()

		Expect(path).To(Equal("fd/f0/blob1/uuid"))
	})
})
