package gcs_test

import (
	"github.com/cloudfoundry-incubator/gcs-blobstore-backup-restore"
	"github.com/cloudfoundry-incubator/gcs-blobstore-backup-restore/fakes"
	. "github.com/onsi/ginkgo"
	//. "github.com/onsi/gomega"
)

var _ = Describe("Restorer", func() {
	var firstBucket *fakes.FakeBucket
	var secondBucket *fakes.FakeBucket
	var firstBackupBucket *fakes.FakeBucket
	var secondBackupBucket *fakes.FakeBucket

	var restorer gcs.Restorer

	const firstBucketName = "first-bucket-name"
	const secondBucketName = "second-bucket-name"
	const firstBackupBucketName = "first-bucket-name"
	const secondBackupBucketName = "second-bucket-name"

	var executionStrategy = gcs.NewParallelStrategy()

	BeforeEach(func() {
		firstBucket = new(fakes.FakeBucket)
		secondBucket = new(fakes.FakeBucket)
		firstBackupBucket = new(fakes.FakeBucket)
		secondBackupBucket = new(fakes.FakeBucket)

		firstBucket.NameReturns(firstBucketName)
		secondBucket.NameReturns(secondBucketName)
		firstBackupBucket.NameReturns(firstBackupBucketName)
		secondBackupBucket.NameReturns(secondBackupBucketName)

		restorer = gcs.NewRestorer(map[string]gcs.BucketPair{
			"first":  {Bucket: firstBucket, BackupBucket: firstBackupBucket},
			"second": {Bucket: secondBucket, BackupBucket: secondBackupBucket},
		}, executionStrategy)
	})

	It("restores the blobs into the live bucket", func() {
		//backups := map[string]gcs.BucketBackup{
		//	"first": {
		//		Name: firstBucketName,
		//		Blobs: []gcs.Blob{
		//			{Name: "blob1", GenerationID: 123},
		//			{Name: "blob2", GenerationID: 234},
		//		},
		//	},
		//	"second": {
		//		Name: secondBucketName,
		//		Blobs: []gcs.Blob{
		//			{Name: "blob3", GenerationID: 345},
		//		},
		//	},
		//}
		//var expectedBlobs []gcs.Blob
		//err := restorer.Restore(backups)
		//
		//Expect(err).NotTo(HaveOccurred())
		//Expect(firstBackupBucket.Copy)
	})
})
