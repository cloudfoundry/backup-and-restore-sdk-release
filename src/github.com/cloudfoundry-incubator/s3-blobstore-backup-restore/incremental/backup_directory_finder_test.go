package incremental_test

import (
	"fmt"

	"github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/incremental"
	"github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/incremental/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Finder", func() {
	Describe("ListBlobs", func() {
		var bucket *fakes.FakeBucket
		var finder *incremental.Finder
		var blobs []incremental.BackedUpBlob
		var err error

		BeforeEach(func() {
			bucket = new(fakes.FakeBucket)
			finder = &incremental.Finder{
				ID:     "bucket_id",
				Bucket: bucket,
			}
		})

		JustBeforeEach(func() {
			blobs, err = finder.ListBlobs()
		})

		Context("when there are no backup directories", func() {
			BeforeEach(func() {
				bucket.ListDirectoriesReturns([]string{"not_a_backup_directory"}, nil)
			})

			It("returns an empty list", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(blobs).To(BeEmpty())
				Expect(bucket.ListBlobsCallCount()).To(BeZero())
			})
		})

		Context("when listing backup directories fails", func() {
			BeforeEach(func() {
				bucket.ListDirectoriesReturns(nil, fmt.Errorf("oups"))
			})

			It("returns an error", func() {
				Expect(err).To(MatchError("oups"))
			})
		})

		Context("when there is a complete backup directory", func() {
			var blob *fakes.FakeBlob

			BeforeEach(func() {
				blob = new(fakes.FakeBlob)
				blob.PathReturns("2000_01_01_01_01_01/bucket_id/f0/fd/blob1/uuid")
				bucket.ListDirectoriesReturns([]string{"2000_01_01_01_01_01"}, nil)
				bucket.ListBlobsReturns([]incremental.Blob{blob}, nil)
				bucket.IsBackupCompleteReturns(true, nil)
			})

			It("returns the list of blobs therein", func() {
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

			Context("and listing the blobs fails", func() {
				BeforeEach(func() {
					bucket.ListBlobsReturns(nil, fmt.Errorf("fail to list"))
				})

				It("returns an error", func() {

					Expect(err).To(MatchError("fail to list"))
				})
			})

		})

		Context("when there are multiple complete backup directories", func() {
			var blob *fakes.FakeBlob

			BeforeEach(func() {
				blob = new(fakes.FakeBlob)
				blob.PathReturns("2000_01_03_01_01_01/bucket_id/f0/fd/blob1/uuid")
				bucket.ListDirectoriesReturns([]string{"2000_01_02_01_01_01", "2000_01_03_01_01_01", "2000_01_01_01_01_01"}, nil)
				bucket.ListBlobsReturns([]incremental.Blob{blob}, nil)
				bucket.IsBackupCompleteReturnsOnCall(0, false, nil)
				bucket.IsBackupCompleteReturnsOnCall(1, true, nil)
			})

			It("returns the list of blobs therein from the latest backup", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(bucket.ListBlobsArgsForCall(0)).To(Equal("2000_01_02_01_01_01/bucket_id"))
			})

			Context("and finding last complete backup fails", func() {
				BeforeEach(func() {
					bucket.IsBackupCompleteReturnsOnCall(0, false, fmt.Errorf("no go"))
				})

				It("returns an error", func() {
					Expect(err).To(MatchError("no go"))
				})
			})
		})

		Context("when all backup directories are incomplete", func() {
			BeforeEach(func() {
				bucket.ListDirectoriesReturns([]string{"2000_01_02_01_01_01"}, nil)
				bucket.IsBackupCompleteReturns(false, nil)
			})
			It("returns an empty list of blobs", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(blobs).To(BeEmpty())
				Expect(bucket.ListBlobsCallCount()).To(BeZero())
			})
		})
	})
})
