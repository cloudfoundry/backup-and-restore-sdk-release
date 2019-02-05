package unversioned_test

import (
	"github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/incremental"
	"github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/incremental/fakes"
	"github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/unversioned"

	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("BucketPair", func() {
	var (
		liveBucket           *fakes.FakeBucket
		backupBucket         *fakes.FakeBucket
		bucketPair           unversioned.RestoreBucketPair
		bucketBackup         incremental.BucketBackup
		err                  error
		configLiveBucket     string
		configLiveRegion     string
		artifactBackupBucket string
		artifactBackupRegion string
	)

	BeforeEach(func() {
		configLiveBucket = "config_live_bucket"
		configLiveRegion = "config_live_region"
		artifactBackupBucket = "artifact_backup_bucket"
		artifactBackupRegion = "artifact_backup_region"

		liveBucket = new(fakes.FakeBucket)
		backupBucket = new(fakes.FakeBucket)
		bucketPair = unversioned.NewRestoreBucketPair(liveBucket, backupBucket)

		liveBucket.NameReturns(configLiveBucket)
		liveBucket.RegionReturns(configLiveRegion)
		backupBucket.NameReturns(artifactBackupBucket)
		backupBucket.RegionReturns(artifactBackupRegion)
	})

	Describe("Restore", func() {
		JustBeforeEach(func() {
			bucketBackup = incremental.BucketBackup{
				BucketName:   artifactBackupBucket,
				BucketRegion: artifactBackupRegion,
				Blobs: []string{
					"2015-12-13-05-06-07/my_bucket_id/livebucketpath/to/real/blob1",
					"2015-12-13-05-06-07/my_bucket_id/livebucketpath/to/real/blob2",
				},
				BackupDirectoryPath: "2015-12-13-05-06-07/my_bucket_id",
			}
			err = bucketPair.Restore(bucketBackup)
		})

		It("successfully copies from the backup bucket to the live bucket", func() {
			By("not returning an error", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			By("copying from the backup location to the live location", func() {
				Expect(liveBucket.CopyBlobFromBucketCallCount()).To(Equal(2))
				bucket, src, dst := liveBucket.CopyBlobFromBucketArgsForCall(0)
				Expect(src).To(Equal("2015-12-13-05-06-07/my_bucket_id/livebucketpath/to/real/blob2"))
				Expect(dst).To(Equal("livebucketpath/to/real/blob2"))
				Expect(bucket.Name()).To(Equal(bucketBackup.BucketName))

				bucket, src, dst = liveBucket.CopyBlobFromBucketArgsForCall(1)
				Expect(src).To(Equal("2015-12-13-05-06-07/my_bucket_id/livebucketpath/to/real/blob1"))
				Expect(dst).To(Equal("livebucketpath/to/real/blob1"))
				Expect(bucket.Name()).To(Equal(bucketBackup.BucketName))
			})
		})

		Context("When CopyObject errors", func() {
			BeforeEach(func() {
				liveBucket.CopyBlobFromBucketReturns(fmt.Errorf("cannot copy object"))
			})

			It("errors", func() {
				Expect(err).To(MatchError(ContainSubstring("cannot copy object")))
			})
		})
	})

	Describe("CheckValidity", func() {
		Context("when the live bucket and the backup bucket are not the same", func() {
			It("returns nil", func() {
				Expect(unversioned.NewRestoreBucketPair(liveBucket, backupBucket).CheckValidity()).To(BeNil())
			})
		})

		Context("when the live bucket and the backup bucket are the same", func() {
			It("returns an error", func() {
				Expect(unversioned.NewRestoreBucketPair(liveBucket, liveBucket).CheckValidity()).To(MatchError("live bucket and backup bucket cannot be the same"))
			})
		})
	})
})
