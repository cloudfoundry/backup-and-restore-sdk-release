package gcs_test

import (
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/gcs-blobstore-backup-restore"
	"github.com/cloudfoundry-incubator/gcs-blobstore-backup-restore/fakes"
)

var _ = Describe("Backuper", func() {
	Describe("CreateLiveBucketSnapshot", func() {
		var bucket *fakes.FakeBucket
		var backupBucket *fakes.FakeBucket

		var backuper gcs.Backuper

		const firstBucketName = "first-bucket-name"

		BeforeEach(func() {
			bucket = new(fakes.FakeBucket)
			bucket.NameReturns(firstBucketName)

			backuper = gcs.NewBackuper(map[string]gcs.BucketPair{
				"first": {
					Bucket:       bucket,
					BackupBucket: backupBucket,
				},
			})
		})

		Context("when there is no previous backup artifact", func() {
			Context("and there is a single bucket to be backed up", func() {
				It("creates a snapshot directory with a copy of the live bucket", func() {
					blob1 := "file_1_a"
					blob2 := "file_1_b"
					bucket.ListBlobsReturns([]gcs.Blob{
						{Name: blob1},
						{Name: blob2},
					}, nil)

					bucket.CopyBlobWithinBucketReturns(0, nil)

					err := backuper.CreateLiveBucketSnapshot()

					Expect(bucket.CopyBlobWithinBucketCallCount()).To(Equal(2))
					Expect(err).NotTo(HaveOccurred())
					blob, path := bucket.CopyBlobWithinBucketArgsForCall(0)
					Expect(blob).To(Equal(blob1))
					Expect(path).To(Equal(fmt.Sprintf("temporary-backup-artifact/%s", blob1)))

					blob, path = bucket.CopyBlobWithinBucketArgsForCall(1)
					Expect(blob).To(Equal(blob2))
					Expect(path).To(Equal(fmt.Sprintf("temporary-backup-artifact/%s", blob2)))
				})
			})
		})

		Context("when there is a previous backup artifact", func() {

		})

		Context("when list blobs fails", func() {
			It("returns an error", func() {
				bucket.ListBlobsReturns(nil, errors.New("ifailed"))
				err := backuper.CreateLiveBucketSnapshot()
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("ifailed"))
			})
		})

		Context("when copy blob fails", func() {
			It("returns an error", func() {
				blob1 := "file_1_a"
				bucket.ListBlobsReturns([]gcs.Blob{
					{Name: blob1},
				}, nil)

				bucket.CopyBlobWithinBucketReturns(0, errors.New("oopsifailed"))
				err := backuper.CreateLiveBucketSnapshot()
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("oopsifailed"))
			})
		})
	})

	Describe("TransferBlobsToBackupBucket", func() {
		var bucket *fakes.FakeBucket
		var backupBucket *fakes.FakeBucket

		var backuper gcs.Backuper

		const firstBucketName = "first-bucket-name"

		BeforeEach(func() {
			bucket = new(fakes.FakeBucket)
			bucket.NameReturns(firstBucketName)

			backupBucket = new(fakes.FakeBucket)
			backupBucket.NameReturns(firstBucketName)

			backuper = gcs.NewBackuper(map[string]gcs.BucketPair{
				"first": {
					Bucket:       bucket,
					BackupBucket: backupBucket,
				},
			})
		})

		Context("when there is no previous backup artifact", func() {
			Context("and there is a single bucket to be backed up", func() {

				It("transfers the blobs from the live bucket to the backup bucket and deletes the blobs from live", func() {
					blob1 := "file_1_a"
					blob2 := "temporary-backup-artifact/file_1_b"
					bucket.ListBlobsReturns([]gcs.Blob{
						{Name: blob1},
						{Name: blob2},
					}, nil)

					bucket.CopyBlobBetweenBucketsReturns(0, nil)

					err := backuper.TransferBlobsToBackupBucket()
					Expect(err).NotTo(HaveOccurred())

					Expect(bucket.CopyBlobBetweenBucketsCallCount()).To(Equal(1))
					dstBucket, blob, path := bucket.CopyBlobBetweenBucketsArgsForCall(0)
					Expect(dstBucket.Name()).To(Equal(backupBucket.Name()))
					Expect(blob).To(Equal(blob2))
					Expect(path).To(Equal(fmt.Sprintf("%s", blob2))) //timestamped dir?

					Expect(bucket.DeleteBlobCallCount()).To(Equal(1))
					Expect(bucket.DeleteBlobArgsForCall(0)).To(Equal(blob2))
				})
			})
		})

		Context("when list blobs fails", func() {
			It("returns an error", func() {
				bucket.ListBlobsReturns(nil, errors.New("ifailed"))
				err := backuper.TransferBlobsToBackupBucket()
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("ifailed"))
			})
		})

		Context("when copy blob fails", func() {
			It("returns an error", func() {
				blob1 := "temporary-backup-artifact/file_1_a"
				bucket.ListBlobsReturns([]gcs.Blob{
					{Name: blob1},
				}, nil)

				bucket.CopyBlobBetweenBucketsReturns(0, errors.New("oopsifailed"))
				err := backuper.TransferBlobsToBackupBucket()
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("oopsifailed"))
			})
		})

		Context("when delete blob fails", func() {

			It("errors", func() {
				blob1 := "file_1_a"
				blob2 := "temporary-backup-artifact/file_1_b"
				bucket.ListBlobsReturns([]gcs.Blob{
					{Name: blob1},
					{Name: blob2},
				}, nil)

				bucket.CopyBlobBetweenBucketsReturns(0, nil)
				bucket.DeleteBlobReturns(errors.New("ifailed"))

				err := backuper.TransferBlobsToBackupBucket()
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("ifailed"))
			})
		})
	})
})
