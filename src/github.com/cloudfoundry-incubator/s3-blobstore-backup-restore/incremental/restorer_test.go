package incremental_test

import (
	"fmt"

	"github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/incremental"
	"github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/incremental/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Restorer", func() {
	var (
		destinationLiveDropletsBucket *fakes.FakeBucket
		sourceBackupDropletsBucket    *fakes.FakeBucket
		destinationLivePackagesBucket *fakes.FakeBucket
		sourceBackupPackagesBucket    *fakes.FakeBucket
		dropletsBucketPair            incremental.RestoreBucketPair
		packagesBucketPair            incremental.RestoreBucketPair
		bucketPairs                   map[string]incremental.RestoreBucketPair
		artifact                      *fakes.FakeArtifact

		err error

		restorer incremental.Restorer

		bucketBackups map[string]incremental.BucketBackup
	)

	BeforeEach(func() {
		destinationLiveDropletsBucket = new(fakes.FakeBucket)
		sourceBackupDropletsBucket = new(fakes.FakeBucket)
		destinationLivePackagesBucket = new(fakes.FakeBucket)
		sourceBackupPackagesBucket = new(fakes.FakeBucket)

		dropletsBucketPair = incremental.NewRestoreBucketPair(destinationLiveDropletsBucket, sourceBackupDropletsBucket)
		packagesBucketPair = incremental.NewRestoreBucketPair(destinationLivePackagesBucket, sourceBackupPackagesBucket)

		artifact = new(fakes.FakeArtifact)
		bucketBackups = map[string]incremental.BucketBackup{
			"droplets": {
				BucketName:          "artifact_backup_droplet_bucket",
				BucketRegion:        "artifact_backup_droplet_region",
				Blobs:               []string{"timestamp/droplets/my_droplet1", "timestamp/droplets/my_droplet2"},
				BackupDirectoryPath: "timestamp/droplets",
			},
			"packages": {
				BucketName:          "artifact_backup_package_bucket",
				BucketRegion:        "artifact_backup_package_region",
				Blobs:               []string{"timestamp/packages/my_package1", "timestamp/packages/my_package2"},
				BackupDirectoryPath: "timestamp/packages",
			},
		}
		artifact.LoadReturns(bucketBackups, nil)

		bucketPairs = map[string]incremental.RestoreBucketPair{
			"droplets": dropletsBucketPair,
			"packages": packagesBucketPair,
		}

		restorer = incremental.NewRestorer(bucketPairs, artifact)
	})

	JustBeforeEach(func() {
		err = restorer.Run()
	})

	Context("When the artifact is valid and copying works", func() {
		BeforeEach(func() {
			artifact.LoadReturns(bucketBackups, nil)
			destinationLiveDropletsBucket.CopyBlobFromBucketReturns(nil)
			destinationLivePackagesBucket.CopyBlobFromBucketReturns(nil)
		})

		It("restores all the bucket pairs", func() {
			Expect(err).NotTo(HaveOccurred())

			srcDroplets := []string{"timestamp/droplets/my_droplet1", "timestamp/droplets/my_droplet2"}
			dstDroplets := []string{"my_droplet1", "my_droplet2"}
			testBucketsWithBlobs(
				destinationLiveDropletsBucket,
				sourceBackupDropletsBucket,
				srcDroplets,
				dstDroplets,
			)

			srcPackages := []string{"timestamp/packages/my_package1", "timestamp/packages/my_package2"}
			dstPackages := []string{"my_package1", "my_package2"}
			testBucketsWithBlobs(
				destinationLivePackagesBucket,
				sourceBackupPackagesBucket,
				srcPackages,
				dstPackages,
			)
		})
	})

	Context("When a bucket cannot be restored", func() {
		BeforeEach(func() {
			destinationLiveDropletsBucket.NameReturns("dropletslivebucket")
			destinationLiveDropletsBucket.CopyBlobFromBucketReturns(fmt.Errorf("restore error1"))
		})

		It("returns an error", func() {
			Expect(err.Error()).To(
				ContainSubstring("failed to restore bucket dropletslivebucket: restore error1\nrestore error1"),
			)
		})
	})

	Context("when the artifact cannot be loaded", func() {
		BeforeEach(func() {
			artifact.LoadReturns(nil, fmt.Errorf("CANNOT LOAD ARTIFACT"))
		})

		It("returns an error", func() {
			Expect(err).To(MatchError("CANNOT LOAD ARTIFACT"))
		})
	})

	Context("When there is a bucket pair in the restore config that is not in the backup artifact", func() {

		BeforeEach(func() {
			notInArtifactBucket1 := new(fakes.FakeBucket)
			notInArtifactBucket2 := new(fakes.FakeBucket)
			notInArtifactPair := incremental.NewRestoreBucketPair(notInArtifactBucket1, notInArtifactBucket2)

			bucketPairs["not-in-artifact"] = notInArtifactPair
			restorer = incremental.NewRestorer(bucketPairs, artifact)
		})

		It("returns an error", func() {
			Expect(err).To(MatchError("cannot restore bucket not-in-artifact, not found in backup artifact"))
		})
	})

	Context("When there is a bucket pair that is recorded to have been empty on backup", func() {

		BeforeEach(func() {
			bucketBackups = map[string]incremental.BucketBackup{
				"droplets": {
					BucketName:          "artifact_backup_droplet_bucket",
					BucketRegion:        "artifact_backup_droplet_region",
					Blobs:               []string{"my_key", "another_key"},
					BackupDirectoryPath: "timestamp/droplets",
				},
				"packages": {
					BucketName:          "artifact_backup_package_bucket",
					BucketRegion:        "artifact_backup_package_region",
					Blobs:               []string{},
					BackupDirectoryPath: "timestamp/packages",
				},
			}
			artifact.LoadReturns(bucketBackups, nil)
		})

		It("does not attempt to restore that pair", func() {
			Expect(destinationLiveDropletsBucket.CopyBlobFromBucketCallCount()).To(Equal(2))
			Expect(destinationLivePackagesBucket.CopyBlobFromBucketCallCount()).To(Equal(0))
		})
	})

	Context("When there is a bucket referenced in the artifact that is not in the restore config", func() {
		BeforeEach(func() {
			bucketBackups["not-in-restore-config"] = incremental.BucketBackup{
				BucketName:          "whatever",
				BucketRegion:        "whatever",
				Blobs:               []string{"timestamp/not-in-restore-config/thing"},
				BackupDirectoryPath: "timestamp/not-in-restore-config",
			}
			artifact.LoadReturns(bucketBackups, nil)
		})

		It("returns an error", func() {
			Expect(err).To(MatchError("restore config does not mention bucket: not-in-restore-config, but is present in the artifact"))
		})
	})
})

func testBucketsWithBlobs(dstLiveBucket *fakes.FakeBucket, srcBackupBucket *fakes.FakeBucket, srcBlobs, dstBlobs []string) {
	var srcBlobSeen []string
	var dstBlobSeen []string

	Expect(dstLiveBucket.CopyBlobFromBucketCallCount()).To(Equal(2))

	bucket, src, dst := dstLiveBucket.CopyBlobFromBucketArgsForCall(0)
	Expect(bucket).To(Equal(srcBackupBucket))
	srcBlobSeen = append(srcBlobSeen, src)
	dstBlobSeen = append(dstBlobSeen, dst)

	bucket, src, dst = dstLiveBucket.CopyBlobFromBucketArgsForCall(1)
	Expect(bucket).To(Equal(srcBackupBucket))
	srcBlobSeen = append(srcBlobSeen, src)
	dstBlobSeen = append(dstBlobSeen, dst)

	Expect(srcBlobSeen).To(ConsistOf(srcBlobs))
	Expect(dstBlobSeen).To(ConsistOf(dstBlobs))
}
