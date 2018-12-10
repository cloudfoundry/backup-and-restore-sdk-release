package gcs_test

import (
	"fmt"

	"github.com/cloudfoundry-incubator/gcs-blobstore-backup-restore"
	"github.com/cloudfoundry-incubator/gcs-blobstore-backup-restore/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Restorer", func() {
	var firstBucket *fakes.FakeBucket
	var secondBucket *fakes.FakeBucket
	var firstBackupBucket *fakes.FakeBucket
	var secondBackupBucket *fakes.FakeBucket

	var restorer gcs.Restorer
	var artifact map[string]gcs.BackupBucketDirectory

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
		config := map[string]gcs.BucketPair{
			"first":  {LiveBucket: firstBucket, BackupBucket: firstBackupBucket},
			"second": {LiveBucket: secondBucket, BackupBucket: secondBackupBucket},
		}

		restorer = gcs.NewRestorer(config)
	})

	Context("when the configuration is valid", func() {
		BeforeEach(func() {
			artifact = map[string]gcs.BackupBucketDirectory{
				"first":  {BucketName: firstBackupBucketName, Path: timestamp + "/first"},
				"second": {BucketName: secondBackupBucketName, Path: timestamp + "/second"},
			}

			firstBackupBucket.CopyBlobsBetweenBucketsReturns(nil)
			secondBackupBucket.CopyBlobsBetweenBucketsReturns(nil)
		})

		It("copies the blobs from the path in the backup bucket to the live bucket for each bucketPair", func() {
			err := restorer.Restore(artifact)
			Expect(err).NotTo(HaveOccurred())

			Expect(firstBackupBucket.CopyBlobsBetweenBucketsCallCount()).To(Equal(1))
			destinationBucket, sourcePath := firstBackupBucket.CopyBlobsBetweenBucketsArgsForCall(0)
			Expect(destinationBucket).To(Equal(firstBucket))
			Expect(sourcePath).To(Equal(timestamp + "/first"))

			Expect(secondBackupBucket.CopyBlobsBetweenBucketsCallCount()).To(Equal(1))
			destinationBucket, sourcePath = secondBackupBucket.CopyBlobsBetweenBucketsArgsForCall(0)
			Expect(destinationBucket).To(Equal(secondBucket))
			Expect(sourcePath).To(Equal(timestamp + "/second"))
		})

		Context("when a copy fails", func() {
			BeforeEach(func() {
				firstBackupBucket.CopyBlobsBetweenBucketsReturns(fmt.Errorf("foo"))
				secondBackupBucket.CopyBlobsBetweenBucketsReturns(fmt.Errorf("foo"))
			})

			It("returns an error on the first failure", func() {
				err := restorer.Restore(artifact)
				Expect(err).To(MatchError("foo"))
			})
		})
	})

	Context("when there is a mismatch between the artifact and the config", func() {
		It("fails if the artifact contains a bucket id that does not exist in the config", func() {
			artifact = map[string]gcs.BackupBucketDirectory{
				"first":  {},
				"second": {},
				"third":  {},
			}
			err := restorer.Restore(artifact)
			Expect(err).To(MatchError("no entry found in restore config for bucket: third"))
		})

		It("fails if the config contains a bucket id that does not exist in the artifact", func() {
			artifact = map[string]gcs.BackupBucketDirectory{"first": {}}
			err := restorer.Restore(artifact)
			Expect(err).To(MatchError("no entry found in restore artifact for bucket: second"))
		})
	})

})
