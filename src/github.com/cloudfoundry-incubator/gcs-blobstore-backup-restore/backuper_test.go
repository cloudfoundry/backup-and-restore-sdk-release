package gcs_test

import (
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/gcs-blobstore-backup-restore"
	"github.com/cloudfoundry-incubator/gcs-blobstore-backup-restore/fakes"
)

var _ = Describe("GCSBackuper", func() {
	Describe("CreateLiveBucketSnapshot", func() {
		var bucket *fakes.FakeBucket
		var backupBucket *fakes.FakeBucket

		var backuper gcs.GCSBackuper

		const firstBucketName = "first-bucket-name"
		const secondBucketName = "second-bucket-name"
		const bucketId = "bucket-id"

		BeforeEach(func() {
			bucket = new(fakes.FakeBucket)
			bucket.NameReturns(firstBucketName)
			backupBucket = new(fakes.FakeBucket)
			backupBucket.NameReturns(secondBucketName)

			backuper = gcs.NewBackuper(map[string]gcs.BucketPair{
				bucketId: {
					Bucket:       bucket,
					BackupBucket: backupBucket,
				},
			})
		})

		Context("when there is no previous backup artifact", func() {
			Context("and there is a single bucket to be backed up", func() {
				var (
					blob1 string
					blob2 string
				)

				BeforeEach(func() {
					blob1 = "file_1_a"
					blob2 = "file_1_b"
					bucket.ListBlobsReturns([]gcs.Blob{
						{Name: blob1},
						{Name: blob2},
					}, nil)

					bucket.CopyBlobWithinBucketReturns(nil)
					bucket.CreateFileReturns(nil)
				})

				It("creates an empty common blobs file", func() {

					_, commonBlobs, err := backuper.CreateLiveBucketSnapshot()
					Expect(err).NotTo(HaveOccurred())

					Expect(commonBlobs[bucketId]).To(BeEmpty())
				})
			})
		})

		Context("when there is a previous backup artifact", func() {
			var blob1, blob2 string
			BeforeEach(func() {
				blob1 = "file_1_a"
				blob2 = "file_1_b"
				bucket.ListBlobsReturns([]gcs.Blob{
					{Name: blob1},
					{Name: blob2},
				}, nil)

				lastBackupBlobs := map[string]gcs.Blob{
					blob1: {Name: "1970_01_01_00_00_00/droplets/" + blob1},
				}

				backupBucket.LastBackupBlobsReturns(lastBackupBlobs, nil)
			})

			It("creates a common blobs file", func() {
				_, commonBlobs, err := backuper.CreateLiveBucketSnapshot()
				Expect(err).NotTo(HaveOccurred())

				Expect(commonBlobs[bucketId]).To(Equal([]gcs.Blob{{Name: "1970_01_01_00_00_00/droplets/" + blob1}}))
			})
		})

		Context("when list blobs fails", func() {
			It("returns an error", func() {
				bucket.ListBlobsReturns(nil, errors.New("ifailed"))
				_, _, err := backuper.CreateLiveBucketSnapshot()
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("ifailed"))
			})
		})

	})

	Describe("CopyBlobsWithinBackupBucket", func() {
		var bucket *fakes.FakeBucket
		var backupBucket *fakes.FakeBucket
		var bucketPairID = "droplets"

		var backuper gcs.GCSBackuper

		const firstBucketName = "first-bucket-name"

		BeforeEach(func() {
			bucket = new(fakes.FakeBucket)
			bucket.NameReturns(firstBucketName)

			backupBucket = new(fakes.FakeBucket)
			backupBucket.NameReturns(firstBucketName)

			backuper = gcs.NewBackuper(map[string]gcs.BucketPair{
				bucketPairID: {
					Bucket:       bucket,
					BackupBucket: backupBucket,
				},
			})
		})

		Context("when all of the blobs are common", func() {
			var blob1 string
			backupBucketAddresses := make(map[string]gcs.BackupBucketDir)
			commonBlobs := make(map[string][]gcs.Blob)

			BeforeEach(func() {
				blob1 = "file1"
				bucket.ListBlobsReturns([]gcs.Blob{
					{Name: blob1},
				}, nil)

				backupBucket.LastBackupBlobsReturns(map[string]gcs.Blob{
					blob1: {Name: "1970_01_01_00_00_00/droplets/" + blob1},
				}, nil)

				backupBucket.CopyBlobBetweenBucketsReturns(nil)
				backupBucketAddresses["droplets"] = gcs.BackupBucketDir{BucketName: firstBucketName, Path: "2006_01_02_15_04_05/droplets"}
				backupBucket.GetBlobReturns([]byte(`[{"name": "1970_01_01_00_00_00/droplets/file1"}]`), nil)
				backupBucket.DeleteBlobReturns(nil)

				commonBlobs[bucketPairID] = []gcs.Blob{{Name: "1970_01_01_00_00_00/droplets/" + blob1}}
			})

			It("copies over all the common blobs from the previous backup", func() {
				err := backuper.CopyBlobsWithinBackupBucket(backupBucketAddresses, commonBlobs)
				Expect(err).NotTo(HaveOccurred())

				Expect(backupBucket.CopyBlobWithinBucketCallCount()).To(Equal(1))
				blob, path := backupBucket.CopyBlobWithinBucketArgsForCall(0)
				Expect(blob).To(Equal("1970_01_01_00_00_00/droplets/" + blob1))
				Expect(path).To(Equal("2006_01_02_15_04_05/" + bucketPairID + "/file1"))
			})
		})

		Context("when the commonBlobs map does not contain a bucket id", func() {
			backupBucketAddresses := make(map[string]gcs.BackupBucketDir)

			BeforeEach(func() {
				backupBucketAddresses["droplets"] = gcs.BackupBucketDir{BucketName: firstBucketName, Path: "2006_01_02_15_04_05/droplets"}
			})

			It("returns an error", func() {
				err := backuper.CopyBlobsWithinBackupBucket(backupBucketAddresses, nil)
				Expect(err).To(MatchError("cannot find commonBlobs for bucket id: droplets"))
			})
		})

		Context("when a common blob is missing", func() {
			backupBucketAddresses := make(map[string]gcs.BackupBucketDir)
			commonBlobs := make(map[string][]gcs.Blob)

			BeforeEach(func() {
				backupBucketAddresses["droplets"] = gcs.BackupBucketDir{BucketName: firstBucketName, Path: "2006_01_02_15_04_05/droplets"}
				backupBucket.GetBlobReturns([]byte(`[{"name": "1970_01_01_00_00_00/droplets/file1"}]`), nil)
				backupBucket.CopyBlobWithinBucketReturns(fmt.Errorf("gcs copy error"))

				commonBlobs["droplets"] = []gcs.Blob{{Name: "heyheyhey"}}
			})

			It("returns the corret error", func() {
				err := backuper.CopyBlobsWithinBackupBucket(backupBucketAddresses, commonBlobs)
				Expect(err).To(MatchError("gcs copy error"))
			})
		})
	})
})
