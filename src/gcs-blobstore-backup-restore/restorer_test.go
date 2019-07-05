package gcs_test

import (
	"fmt"

	"gcs-blobstore-backup-restore"
	"gcs-blobstore-backup-restore/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Restorer", func() {
	var firstBucket *fakes.FakeBucket
	var secondBucket *fakes.FakeBucket
	var firstBackupBucket *fakes.FakeBucket
	var secondBackupBucket *fakes.FakeBucket

	var restorer gcs.Restorer
	var artifact map[string]gcs.BucketBackup

	const firstBucketName = "first-bucket-name"
	const secondBucketName = "second-bucket-name"
	const firstBackupBucketName = "first-bucket-name"
	const secondBackupBucketName = "second-bucket-name"
	const timestamp = "2006_01_02_15_04_05"

	BeforeEach(func() {
		firstBucket = new(fakes.FakeBucket)
		secondBucket = new(fakes.FakeBucket)
		firstBackupBucket = new(fakes.FakeBucket)
		secondBackupBucket = new(fakes.FakeBucket)

		firstBucket.NameReturns(firstBucketName)
		secondBucket.NameReturns(secondBucketName)
		firstBackupBucket.NameReturns(firstBackupBucketName)
		secondBackupBucket.NameReturns(secondBackupBucketName)
		config := map[string]gcs.BackupToComplete{
			"first": {
				BucketPair:     gcs.BucketPair{LiveBucket: firstBucket},
				SameAsBucketID: "",
			},
			"second": {
				BucketPair:     gcs.BucketPair{LiveBucket: secondBucket},
				SameAsBucketID: "",
			},
		}

		restorer = gcs.NewRestorer(config)
	})

	Context("when the configuration is valid", func() {
		BeforeEach(func() {
			artifact = map[string]gcs.BucketBackup{
				"first":  {BucketName: firstBackupBucketName, Bucket: firstBackupBucket, Path: timestamp + "/first"},
				"second": {BucketName: secondBackupBucketName, Bucket: secondBackupBucket, Path: timestamp + "/second"},
			}

			firstBackupBucket.CopyBlobsToBucketReturns(nil)
			secondBackupBucket.CopyBlobsToBucketReturns(nil)
		})

		It("copies the blobs from the path in the backup bucket to the live bucket for each bucketPair", func() {
			err := restorer.Restore(artifact)
			Expect(err).NotTo(HaveOccurred())

			Expect(firstBackupBucket.CopyBlobsToBucketCallCount()).To(Equal(1))
			destinationBucket, sourcePath := firstBackupBucket.CopyBlobsToBucketArgsForCall(0)
			Expect(destinationBucket).To(Equal(firstBucket))
			Expect(sourcePath).To(Equal(timestamp + "/first"))

			Expect(secondBackupBucket.CopyBlobsToBucketCallCount()).To(Equal(1))
			destinationBucket, sourcePath = secondBackupBucket.CopyBlobsToBucketArgsForCall(0)
			Expect(destinationBucket).To(Equal(secondBucket))
			Expect(sourcePath).To(Equal(timestamp + "/second"))
		})

		Context("when the live buckets in the configuration point to the same bucket", func() {
			BeforeEach(func() {
				config := map[string]gcs.BackupToComplete{
					"first": {
						BucketPair:     gcs.BucketPair{LiveBucket: firstBucket},
						SameAsBucketID: "",
					},
					"second": {
						BucketPair:     gcs.BucketPair{LiveBucket: firstBackupBucket},
						SameAsBucketID: "",
					},
				}

				artifact = map[string]gcs.BucketBackup{
					"first":  {BucketName: firstBackupBucketName, Bucket: firstBackupBucket, Path: timestamp + "/first"},
					"second": {SameBucketAs: "first"},
				}

				restorer = gcs.NewRestorer(config)
			})

			It("copies the blobs from the path in the backup bucket to the live bucket only once", func() {
				err := restorer.Restore(artifact)
				Expect(err).NotTo(HaveOccurred())

				Expect(firstBackupBucket.CopyBlobsToBucketCallCount()).To(Equal(1))
				destinationBucket, sourcePath := firstBackupBucket.CopyBlobsToBucketArgsForCall(0)
				Expect(destinationBucket).To(Equal(firstBucket))
				Expect(sourcePath).To(Equal(timestamp + "/first"))
			})
		})

		Context("when a copy fails", func() {
			BeforeEach(func() {
				firstBackupBucket.CopyBlobsToBucketReturns(fmt.Errorf("foo"))
				secondBackupBucket.CopyBlobsToBucketReturns(fmt.Errorf("foo"))
			})

			It("returns an error on the first failure", func() {
				err := restorer.Restore(artifact)
				Expect(err).To(MatchError("foo"))
			})
		})
	})

	Context("when there is a mismatch between the artifact and the config", func() {
		It("fails if the artifact contains a bucket id that does not exist in the config", func() {
			artifact = map[string]gcs.BucketBackup{
				"first":  {},
				"second": {},
				"third":  {},
			}
			err := restorer.Restore(artifact)
			Expect(err).To(MatchError("no entry found in restore config for bucket: third"))
		})

		It("fails if the config contains a bucket id that does not exist in the artifact", func() {
			artifact = map[string]gcs.BucketBackup{"first": {}}
			err := restorer.Restore(artifact)
			Expect(err).To(MatchError("no entry found in restore artifact for bucket: second"))
		})
	})

})
