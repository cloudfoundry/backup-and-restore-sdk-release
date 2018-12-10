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
		var liveBucket *fakes.FakeBucket
		var backupBucket *fakes.FakeBucket
		var backuper gcs.Backuper
		var artifactFinder *fakes.FakeBackupArtifactFinder

		var blob1, blob2 string

		const LiveBucketName = "first-bucket-name"
		const BackupBucketName = "second-bucket-name"
		const bucketId = "bucket-id"

		BeforeEach(func() {
			liveBucket = new(fakes.FakeBucket)
			liveBucket.NameReturns(LiveBucketName)
			backupBucket = new(fakes.FakeBucket)
			backupBucket.NameReturns(BackupBucketName)
			artifactFinder = new(fakes.FakeBackupArtifactFinder)
			blob1 = "file_1_a"
			blob2 = "file_1_b"

			backuper = gcs.NewBackuper(map[string]gcs.BucketPair{
				bucketId: {
					LiveBucket:   liveBucket,
					BackupBucket: backupBucket,
					BackupFinder: artifactFinder,
				},
			})
		})

		Context("when there is no previous backup artifact", func() {

			BeforeEach(func() {
				liveBucket.ListBlobsReturns([]gcs.Blob{
					{Name: blob1},
					{Name: blob2},
				}, nil)

				liveBucket.CopyBlobWithinBucketReturns(nil)
			})

			It("returns an empty common blobs map", func() {
				backupBuckets, commonBlobs, err := backuper.CreateLiveBucketSnapshot()
				Expect(err).NotTo(HaveOccurred())

				Expect(commonBlobs[bucketId]).To(BeEmpty())
				Expect(backupBuckets[bucketId].BucketName).To(Equal(BackupBucketName))
				Expect(backupBuckets[bucketId].Path).To(MatchRegexp(".*\\d{4}_\\d{2}_\\d{2}_\\d{2}_\\d{2}_\\d{2}/.*"))

			})

		})

		Context("when there is a previous backup artifact", func() {
			BeforeEach(func() {
				liveBucket.ListBlobsReturns([]gcs.Blob{
					{Name: blob1},
					{Name: blob2},
				}, nil)

				lastBackupBlobs := map[string]gcs.Blob{
					blob1: {Name: "1970_01_01_00_00_00/droplets/" + blob1},
				}

				artifactFinder.ListBlobsReturns(lastBackupBlobs, nil)
			})

			It("returns a map of common blobs", func() {
				_, commonBlobs, err := backuper.CreateLiveBucketSnapshot()

				Expect(err).NotTo(HaveOccurred())
				Expect(commonBlobs[bucketId]).To(Equal([]gcs.Blob{{Name: "1970_01_01_00_00_00/droplets/" + blob1}}))
			})

			It("returns a map of valid BackupBucketDirectory", func() {
				backupBucketDir, _, err := backuper.CreateLiveBucketSnapshot()
				Expect(err).NotTo(HaveOccurred())

				Expect(backupBucketDir[bucketId].BucketName).To(Equal(backupBucket.Name()))
				Expect(backupBucketDir[bucketId].Path).To(MatchRegexp(".*\\d{4}_\\d{2}_\\d{2}_\\d{2}_\\d{2}_\\d{2}/.*"))
			})
		})

		Context("when listing LastBackupBlobs from backup bucket fails", func() {
			It("returns an error", func() {
				artifactFinder.ListBlobsReturns(nil, errors.New("i failed to return last backup blobs"))
				_, _, err := backuper.CreateLiveBucketSnapshot()
				Expect(err).To(MatchError("i failed to return last backup blobs"))
			})
		})

		Context("when list blobs fails", func() {
			It("returns an error", func() {
				liveBucket.ListBlobsReturns(nil, errors.New("ifailed"))
				_, _, err := backuper.CreateLiveBucketSnapshot()
				Expect(err).To(MatchError("ifailed"))
			})
		})

		Context("when copy blob to backup bucket fails", func() {
			BeforeEach(func() {
				liveBucket.ListBlobsReturns([]gcs.Blob{
					{Name: blob1}}, nil)
			})

			It("returns an error", func() {
				liveBucket.CopyBlobBetweenBucketsReturns(errors.New("i failed to copy blob2 to backup bucket"))
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
					LiveBucket:   bucket,
					BackupBucket: backupBucket,
				},
			})
		})

		Context("when all of the blobs are common", func() {
			var blob1 string
			backupBucketAddresses := make(map[string]gcs.BackupBucketDirectory)
			commonBlobs := make(map[string][]gcs.Blob)

			BeforeEach(func() {
				blob1 = "file1"
				bucket.ListBlobsReturns([]gcs.Blob{
					{Name: blob1},
				}, nil)

				backupBucket.LastBackupBlobsReturns(map[string]gcs.Blob{
					blob1: {Name: fmt.Sprintf("1970_01_01_00_00_00/%s/%s", bucketPairID, blob1)},
				}, nil)

				backupBucket.CopyBlobBetweenBucketsReturns(nil)
				backupBucketAddresses["droplets"] = gcs.BackupBucketDirectory{BucketName: firstBucketName, Path: "2006_01_02_15_04_05/" + bucketPairID}
				backupBucket.DeleteBlobReturns(nil)

				commonBlobs[bucketPairID] = []gcs.Blob{{Name: fmt.Sprintf("1970_01_01_00_00_00/%s/%s", bucketPairID, blob1)}}
			})

			It("copies over all the common blobs from the previous backup", func() {
				err := backuper.CopyBlobsWithinBackupBucket(backupBucketAddresses, commonBlobs)
				Expect(err).NotTo(HaveOccurred())

				Expect(backupBucket.CopyBlobWithinBucketCallCount()).To(Equal(1))
				blob, path := backupBucket.CopyBlobWithinBucketArgsForCall(0)
				Expect(blob).To(Equal(fmt.Sprintf("1970_01_01_00_00_00/%s/%s", bucketPairID, blob1)))
				Expect(path).To(Equal("2006_01_02_15_04_05/" + bucketPairID + "/file1"))
			})

			It("writes a backup complete file", func() {
				err := backuper.CopyBlobsWithinBackupBucket(backupBucketAddresses, commonBlobs)
				Expect(err).NotTo(HaveOccurred())

				Expect(backupBucket.CreateBackupCompleteBlobCallCount()).To(Equal(1))
				Expect(backupBucket.CreateBackupCompleteBlobArgsForCall(0)).To(Equal("2006_01_02_15_04_05/" + bucketPairID))
			})
		})

		Context("when the commonBlobs map does not contain a bucket id", func() {
			backupBucketAddresses := make(map[string]gcs.BackupBucketDirectory)

			BeforeEach(func() {
				backupBucketAddresses["droplets"] = gcs.BackupBucketDirectory{BucketName: firstBucketName, Path: "2006_01_02_15_04_05/droplets"}
			})

			It("returns an error", func() {
				err := backuper.CopyBlobsWithinBackupBucket(backupBucketAddresses, nil)
				Expect(err).To(MatchError("cannot find commonBlobs for bucket id: droplets"))

				Expect(backupBucket.CreateBackupCompleteBlobCallCount()).To(BeZero())
			})
		})

		Context("when a common blob is missing", func() {
			backupBucketAddresses := make(map[string]gcs.BackupBucketDirectory)
			commonBlobs := make(map[string][]gcs.Blob)

			BeforeEach(func() {
				backupBucketAddresses["droplets"] = gcs.BackupBucketDirectory{BucketName: firstBucketName, Path: "2006_01_02_15_04_05/droplets"}
				backupBucket.CopyBlobWithinBucketReturns(fmt.Errorf("gcs copy error"))

				commonBlobs["droplets"] = []gcs.Blob{{Name: "heyheyhey"}}
			})

			It("returns the correct error", func() {
				err := backuper.CopyBlobsWithinBackupBucket(backupBucketAddresses, commonBlobs)
				Expect(err).To(MatchError("gcs copy error"))

				Expect(backupBucket.CreateBackupCompleteBlobCallCount()).To(BeZero())
			})
		})

		Context("when creating a backup complete blob fails", func() {
			It("returns the correct error", func() {
				backupBucket.CreateBackupCompleteBlobReturns(errors.New("fail"))
				backuper = gcs.NewBackuper(map[string]gcs.BucketPair{
					bucketPairID: {
						LiveBucket:   bucket,
						BackupBucket: backupBucket,
					},
				})

				backupBucketAddresses := map[string]gcs.BackupBucketDirectory{
					"droplets": {BucketName: firstBucketName, Path: "2006_01_02_15_04_05/" + bucketPairID},
				}
				commonBlobs := map[string][]gcs.Blob{"droplets": {}}
				err := backuper.CopyBlobsWithinBackupBucket(backupBucketAddresses, commonBlobs)

				Expect(err).To(MatchError("fail"))
			})
		})
	})
})
