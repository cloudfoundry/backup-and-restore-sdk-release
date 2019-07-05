package incremental_test

import (
	"fmt"

	"s3-blobstore-backup-restore/incremental"
	"s3-blobstore-backup-restore/incremental/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("BackupBucketPair", func() {
	var (
		liveBucket         *fakes.FakeBucket
		backupBucket       *fakes.FakeBucket
		liveBlob1          *fakes.FakeBlob
		liveBlob2          *fakes.FakeBlob
		bucketPair         incremental.BackupBucketPair
		err                error
		existingBlobs      []incremental.BackedUpBlob
		configLiveBucket   string
		configLiveRegion   string
		configBackupBucket string
		configBackupRegion string
	)

	BeforeEach(func() {
		configLiveBucket = "config_live_bucket"
		configLiveRegion = "config_live_region"
		configBackupBucket = "config_backup_bucket"
		configBackupRegion = "config_backup_bucket"

		liveBucket = new(fakes.FakeBucket)
		backupBucket = new(fakes.FakeBucket)
		liveBlob1 = new(fakes.FakeBlob)
		liveBlob2 = new(fakes.FakeBlob)
		bucketPair = incremental.BackupBucketPair{
			ConfigLiveBucket:   liveBucket,
			ConfigBackupBucket: backupBucket,
		}

		liveBucket.NameReturns(configLiveBucket)
		liveBucket.RegionReturns(configLiveRegion)
		backupBucket.NameReturns(configBackupBucket)
		backupBucket.RegionReturns(configBackupRegion)
	})

	Describe("CopyNewLiveBlobsToBackup", func() {
		BeforeEach(func() {
			liveBlob1.PathReturns("livebucketpath/to/real/blob1")
			liveBlob2.PathReturns("livebucketpath/to/real/blob2")
		})

		It("successfully copies from the live bucket to the backup bucket", func() {
			existingBlobs, err = bucketPair.CopyNewLiveBlobsToBackup(
				[]incremental.BackedUpBlob{
					{
						Path:                "2015-12-13-05-06-07/my_bucket_id/livebucketpath/to/real/blob1",
						BackupDirectoryPath: "2015-12-13-05-06-07/my_bucket_id",
					},
					{
						Path:                "2015-12-13-05-06-07/my_bucket_id/livebucketpath/to/real/blob3",
						BackupDirectoryPath: "2015-12-13-05-06-07/my_bucket_id",
					},
				},
				[]incremental.Blob{liveBlob1, liveBlob2},
				"2015-12-13-05-06-07/my_bucket_id",
			)
			Expect(err).NotTo(HaveOccurred())
			Expect(backupBucket.CopyBlobFromBucketCallCount()).To(Equal(1))
			srcBucket, srcPath, dstPath := backupBucket.CopyBlobFromBucketArgsForCall(0)
			Expect(srcBucket.Name()).To(Equal(configLiveBucket))
			Expect(srcPath).To(Equal("livebucketpath/to/real/blob2"))
			Expect(dstPath).To(Equal("2015-12-13-05-06-07/my_bucket_id/livebucketpath/to/real/blob2"))

			Expect(existingBlobs).To(Equal([]incremental.BackedUpBlob{
				{
					Path:                "2015-12-13-05-06-07/my_bucket_id/livebucketpath/to/real/blob1",
					BackupDirectoryPath: "2015-12-13-05-06-07/my_bucket_id",
				},
			}))
		})

		Context("When CopyObject errors", func() {
			It("errors", func() {
				backupBucket.CopyBlobFromBucketReturns(fmt.Errorf("cannot copy object"))
				_, err = bucketPair.CopyNewLiveBlobsToBackup([]incremental.BackedUpBlob{}, []incremental.Blob{liveBlob1}, "")
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(ContainSubstring("cannot copy object")))
			})
		})
	})
})
