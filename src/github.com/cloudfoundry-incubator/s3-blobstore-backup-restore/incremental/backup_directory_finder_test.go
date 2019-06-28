package incremental_test

import (
	"fmt"

	"github.com/cloudfoundry-incubator/backup-and-restore-sdk/s3-blobstore-backup-restore/incremental"
	"github.com/cloudfoundry-incubator/backup-and-restore-sdk/s3-blobstore-backup-restore/incremental/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Finder", func() {
	Describe("ListBlobs", func() {
		const bucketID = "bucket_id"
		var bucket *fakes.FakeBucket
		var finder incremental.Finder

		BeforeEach(func() {
			bucket = new(fakes.FakeBucket)
			finder = incremental.Finder{}
		})

		Context("when there are no backup directories", func() {
			BeforeEach(func() {
				bucket.ListDirectoriesReturns([]string{"not_a_backup_directory"}, nil)
			})

			It("returns an empty list", func() {
				blobs, err := finder.ListBlobs(bucketID, bucket)

				Expect(err).NotTo(HaveOccurred())
				Expect(blobs).To(BeEmpty())
				Expect(bucket.ListBlobsCallCount()).To(BeZero())
			})
		})

		Context("when all backup directories are incomplete", func() {
			It("returns an empty list of blobs", func() {
				bucket.ListDirectoriesReturns([]string{"2000_01_02_01_01_01"}, nil)
				bucket.HasBlobReturns(false, nil)

				blobs, err := finder.ListBlobs(bucketID, bucket)

				Expect(err).NotTo(HaveOccurred())
				Expect(blobs).To(BeEmpty())
				Expect(bucket.ListBlobsCallCount()).To(BeZero())
			})
		})

		Context("when there is a complete backup directory", func() {
			BeforeEach(func() {
				blob := new(fakes.FakeBlob)
				blob.PathReturns("f0/fd/blob1/uuid")

				bucket.ListDirectoriesReturns([]string{"2000_01_01_01_01_01"}, nil)
				bucket.ListBlobsReturns([]incremental.Blob{blob}, nil)
				bucket.HasBlobReturns(true, nil)
			})

			It("returns the list of blobs therein", func() {
				blobs, err := finder.ListBlobs(bucketID, bucket)

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
				It("returns an error", func() {
					bucket.ListBlobsReturns(nil, fmt.Errorf("fail to list"))

					_, err := finder.ListBlobs(bucketID, bucket)

					Expect(err).To(MatchError("fail to list"))
				})
			})
		})

		Context("when there are multiple complete backup directories", func() {
			BeforeEach(func() {
				blob := new(fakes.FakeBlob)
				blob.PathReturns("f0/fd/blob1/uuid")

				bucket.ListDirectoriesReturns([]string{
					"2000_01_02_01_01_01",
					"2000_01_03_01_01_01",
					"2000_01_01_01_01_01",
				}, nil)
				bucket.ListBlobsReturns([]incremental.Blob{blob}, nil)
			})

			It("returns the list of blobs therein from the latest backup", func() {
				bucket.HasBlobReturnsOnCall(0, false, nil)
				bucket.HasBlobReturnsOnCall(1, false, nil)
				bucket.HasBlobReturnsOnCall(2, true, nil)

				blobs, err := finder.ListBlobs(bucketID, bucket)

				Expect(err).NotTo(HaveOccurred())
				Expect(bucket.ListBlobsArgsForCall(0)).To(Equal("2000_01_01_01_01_01/bucket_id"))
				Expect(blobs).To(ConsistOf(incremental.BackedUpBlob{
					Path:                "2000_01_01_01_01_01/bucket_id/f0/fd/blob1/uuid",
					BackupDirectoryPath: "2000_01_01_01_01_01/bucket_id",
				}))
			})

			Context("and finding last complete backup fails", func() {
				It("returns an error", func() {
					bucket.HasBlobReturns(false, fmt.Errorf("no go"))

					_, err := finder.ListBlobs(bucketID, bucket)

					Expect(err).To(MatchError(ContainSubstring("no go")))
				})
			})
		})

		Context("when listing backup directories fails", func() {
			It("returns an error", func() {
				bucket.ListDirectoriesReturns(nil, fmt.Errorf("oups"))

				_, err := finder.ListBlobs(bucketID, bucket)

				Expect(err).To(MatchError("oups"))
			})
		})
	})
})
