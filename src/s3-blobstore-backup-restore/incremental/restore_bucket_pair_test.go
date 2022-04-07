package incremental_test

import (
	"github.com/cloudfoundry/backup-and-restore-sdk-release/src/s3-blobstore-backup-restore/incremental"
	"github.com/cloudfoundry/backup-and-restore-sdk-release/src/s3-blobstore-backup-restore/incremental/fakes"

	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RestoreBucketPair", func() {
	var (
		liveBucket           *fakes.FakeBucket
		backupBucket         *fakes.FakeBucket
		bucketPair           incremental.RestoreBucketPair
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
		bucketPair = incremental.RestoreBucketPair{
			ConfigLiveBucket:     liveBucket,
			ArtifactBackupBucket: backupBucket,
		}

		liveBucket.NameReturns(configLiveBucket)
		liveBucket.RegionReturns(configLiveRegion)
		backupBucket.NameReturns(artifactBackupBucket)
		backupBucket.RegionReturns(artifactBackupRegion)
	})

	Describe("Restore", func() {
		var backup incremental.Backup

		BeforeEach(func() {
			backup = incremental.Backup{
				BucketName:   artifactBackupBucket,
				BucketRegion: artifactBackupRegion,
				Blobs: []string{
					"2015-12-13-05-06-07/my_bucket_id/livebucketpath/to/real/blob1",
					"2015-12-13-05-06-07/my_bucket_id/livebucketpath/to/real/blob2",
				},
				SrcBackupDirectoryPath: "2015-12-13-05-06-07/my_bucket_id",
			}
		})

		It("successfully copies from the backup bucket to the live bucket", func() {
			err = bucketPair.Restore(backup)

			Expect(err).NotTo(HaveOccurred())
			Expect(liveBucket.CopyBlobFromBucketCallCount()).To(Equal(2))
			bucket0, src0, dst0 := liveBucket.CopyBlobFromBucketArgsForCall(0)
			bucket1, src1, dst1 := liveBucket.CopyBlobFromBucketArgsForCall(1)
			Expect([][]string{{bucket0.Name(), src0, dst0}, {bucket1.Name(), src1, dst1}}).To(ConsistOf(
				[]string{
					backup.BucketName,
					"2015-12-13-05-06-07/my_bucket_id/livebucketpath/to/real/blob1",
					"livebucketpath/to/real/blob1",
				},
				[]string{
					backup.BucketName,
					"2015-12-13-05-06-07/my_bucket_id/livebucketpath/to/real/blob2",
					"livebucketpath/to/real/blob2",
				},
			))
		})

		Context("When CopyObject errors", func() {
			It("errors", func() {
				liveBucket.CopyBlobFromBucketReturns(fmt.Errorf("cannot copy object"))
				err = bucketPair.Restore(backup)
				Expect(err).To(MatchError(ContainSubstring("cannot copy object")))
			})
		})
	})
})
