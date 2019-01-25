package incremental_test

import (
	"github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/incremental"
	"github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/incremental/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("BackupDirectoryFinder", func() {
	Context("when there are no backup directories", func() {
		It("returns an empty list", func() {
			bucket := new(fakes.FakeBucket)
			bucket.ListDirectoriesReturns([]string{"not_a_backup_directory"}, nil)
			finder := incremental.BackupDirectoryFinder{
				Bucket: bucket,
			}

			blobs, err := finder.ListBlobs("bucket_id")

			Expect(err).NotTo(HaveOccurred())
			Expect(blobs).To(BeEmpty())
			Expect(bucket.ListBlobsCallCount()).To(BeZero())
		})
	})

	Context("when there is a complete backup directory", func() {
		It("returns the list of blobs therein", func() {
			blob := new(fakes.FakeBlob)
			blob.PathReturns("2000_01_01_01_01_01/bucket_id/f0/fd/blob1/uuid")
			bucket := new(fakes.FakeBucket)
			bucket.ListDirectoriesReturns([]string{"2000_01_01_01_01_01"}, nil)
			bucket.ListBlobsReturns([]incremental.Blob{blob}, nil)
			finder := incremental.BackupDirectoryFinder{
				ID:     "bucket_id",
				Bucket: bucket,
			}

			blobs, err := finder.ListBlobs("bucket_id")

			Expect(err).NotTo(HaveOccurred())
			Expect(bucket.ListDirectoriesCallCount()).To(Equal(1))
			Expect(bucket.ListBlobsCallCount()).To(Equal(1))
			Expect(bucket.ListBlobsArgsForCall(0)).To(Equal("2000_01_01_01_01_01/bucket_id"))
			Expect(blobs).To(ConsistOf(
				incremental.BackedUpBlob{
					Path:                "2000_01_01_01_01_01/bucket_id/f0/fd/blob1/uuid",
					BackupDirectoryPath: "2000_01_01_01_01_01/bucket_id",
				}))
		})
	})
})
