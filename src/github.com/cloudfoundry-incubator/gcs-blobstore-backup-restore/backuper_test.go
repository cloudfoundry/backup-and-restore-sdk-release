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

		var blob1, blob2 string

		const firstBucketName = "first-bucket-name"
		const secondBucketName = "second-bucket-name"
		const bucketId = "bucket-id"

		BeforeEach(func() {
			bucket = new(fakes.FakeBucket)
			bucket.NameReturns(firstBucketName)
			backupBucket = new(fakes.FakeBucket)
			backupBucket.NameReturns(secondBucketName)
			blob1 = "file_1_a"
			blob2 = "file_1_b"

			backuper = gcs.NewBackuper(map[string]gcs.BucketPair{
				bucketId: {
					Bucket:       bucket,
					BackupBucket: backupBucket,
				},
			})
		})

		Context("when there is no previous backup artifact", func() {

			BeforeEach(func() {

				bucket.ListBlobsReturns([]gcs.Blob{
					{Name: blob1},
					{Name: blob2},
				}, nil)

				bucket.CopyBlobWithinBucketReturns(nil)
			})

			It("returns an empty common blobs map", func() {
				_, commonBlobs, err := backuper.CreateLiveBucketSnapshot()
				Expect(err).NotTo(HaveOccurred())

				Expect(commonBlobs[bucketId]).To(BeEmpty())
			})

		})

		Context("when there is a previous backup artifact", func() {
			BeforeEach(func() {
				bucket.ListBlobsReturns([]gcs.Blob{
					{Name: blob1},
					{Name: blob2},
				}, nil)

				lastBackupBlobs := map[string]gcs.Blob{
					blob1: {Name: "1970_01_01_00_00_00/droplets/" + blob1},
				}

				backupBucket.LastBackupBlobsReturns(lastBackupBlobs, nil)
			})

			It("returns a map of common blobs", func() {
				_, commonBlobs, err := backuper.CreateLiveBucketSnapshot()
				Expect(err).NotTo(HaveOccurred())

				Expect(commonBlobs[bucketId]).To(Equal([]gcs.Blob{{Name: "1970_01_01_00_00_00/droplets/" + blob1}}))
			})

			It("returns a map of valid BackupBucketDir", func() {
				backupBucketDir, _, err := backuper.CreateLiveBucketSnapshot()
				Expect(err).NotTo(HaveOccurred())

				Expect(backupBucketDir[bucketId].BucketName).To(Equal(backupBucket.Name()))
				Expect(backupBucketDir[bucketId].Path).To(MatchRegexp(".*\\d{4}_\\d{2}_\\d{2}_\\d{2}_\\d{2}_\\d{2}/.*"))
			})
		})

		Context("when listing LastBackupBlobs from backup bucket fails", func() {
			It("returns an error", func() {
				backupBucket.LastBackupBlobsReturns(nil, errors.New("i failed to return last backup blobs"))
				_, _, err := backuper.CreateLiveBucketSnapshot()
				Expect(err).To(MatchError("i failed to return last backup blobs"))
			})
		})

		Context("when list blobs fails", func() {
			It("returns an error", func() {
				bucket.ListBlobsReturns(nil, errors.New("ifailed"))
				_, _, err := backuper.CreateLiveBucketSnapshot()
				Expect(err).To(MatchError("ifailed"))
			})
		})

		Context("when copy blob to backup bucket fails", func() {
			BeforeEach(func() {
				bucket.ListBlobsReturns([]gcs.Blob{
					{Name: blob1}}, nil)
			})

			It("returns an error", func() {
				bucket.CopyBlobBetweenBucketsReturns(errors.New("i failed to copy blob2 to backup bucket"))
				_, _, err := backuper.CreateLiveBucketSnapshot()
				Expect(err).To(MatchError("i failed to copy blob2 to backup bucket"))
			})
		})

	})

	Describe("CopyBlobsWithinBackupBucket", func() {
		var bucket *fakes.FakeBucket
		var backupBucket *fakes.FakeBucket
		var bucketPairID = "droplets"

		var backuper gcs.Backuper

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
