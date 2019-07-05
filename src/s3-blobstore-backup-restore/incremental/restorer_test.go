package incremental_test

import (
	"fmt"

	"s3-blobstore-backup-restore/s3bucket"

	"s3-blobstore-backup-restore/incremental"
	"s3-blobstore-backup-restore/incremental/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Restorer", func() {

	const (
		dropletsBlob1 = "timestamp/droplets/my_droplet1"
		dropletsBlob2 = "timestamp/droplets/my_droplet2"
		packagesBlob1 = "timestamp/packages/my_package1"
		packagesBlob2 = "timestamp/packages/my_package2"
	)
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

		backups map[string]incremental.Backup
	)

	BeforeEach(func() {
		destinationLiveDropletsBucket = new(fakes.FakeBucket)
		sourceBackupDropletsBucket = new(fakes.FakeBucket)
		destinationLivePackagesBucket = new(fakes.FakeBucket)
		sourceBackupPackagesBucket = new(fakes.FakeBucket)

		dropletsBucketPair = incremental.RestoreBucketPair{
			ConfigLiveBucket:     destinationLiveDropletsBucket,
			ArtifactBackupBucket: sourceBackupDropletsBucket,
		}
		packagesBucketPair = incremental.RestoreBucketPair{
			ConfigLiveBucket:     destinationLivePackagesBucket,
			ArtifactBackupBucket: sourceBackupPackagesBucket,
		}

		artifact = new(fakes.FakeArtifact)
		backups = map[string]incremental.Backup{
			"droplets": {
				BucketName:             "artifact_backup_droplet_bucket",
				BucketRegion:           "artifact_backup_droplet_region",
				Blobs:                  []string{dropletsBlob1, dropletsBlob2},
				SrcBackupDirectoryPath: "timestamp/droplets",
			},
			"packages": {
				BucketName:             "artifact_backup_package_bucket",
				BucketRegion:           "artifact_backup_package_region",
				Blobs:                  []string{packagesBlob1, packagesBlob2},
				SrcBackupDirectoryPath: "timestamp/packages",
			},
		}

		sourceBackupDropletsBucket.ListBlobsReturns([]incremental.Blob{
			s3bucket.NewBlob("my_droplet1"),
			s3bucket.NewBlob("my_droplet2"),
		}, nil)

		sourceBackupPackagesBucket.ListBlobsReturns([]incremental.Blob{
			s3bucket.NewBlob("my_package1"),
			s3bucket.NewBlob("my_package2"),
		}, nil)
		artifact.LoadReturns(backups, nil)

		bucketPairs = map[string]incremental.RestoreBucketPair{
			"droplets": dropletsBucketPair,
			"packages": packagesBucketPair,
		}
	})

	JustBeforeEach(func() {
		restorer := incremental.NewRestorer(bucketPairs, artifact)
		err = restorer.Run()
	})

	Context("When the artifact is valid and copying works", func() {
		BeforeEach(func() {
			artifact.LoadReturns(backups, nil)
			destinationLiveDropletsBucket.CopyBlobFromBucketReturns(nil)
			destinationLivePackagesBucket.CopyBlobFromBucketReturns(nil)
		})

		It("restores all the bucket pairs", func() {
			Expect(err).NotTo(HaveOccurred())

			srcDroplets := []string{dropletsBlob1, dropletsBlob2}
			dstDroplets := []string{"my_droplet1", "my_droplet2"}
			testBucketsWithBlobs(
				destinationLiveDropletsBucket,
				sourceBackupDropletsBucket,
				srcDroplets,
				dstDroplets,
			)

			srcPackages := []string{packagesBlob1, packagesBlob2}
			dstPackages := []string{"my_package1", "my_package2"}
			testBucketsWithBlobs(
				destinationLivePackagesBucket,
				sourceBackupPackagesBucket,
				srcPackages,
				dstPackages,
			)
		})
	})

	Context("when a bucket is marked same as another", func() {
		BeforeEach(func() {
			backups = map[string]incremental.Backup{
				"droplets": {
					BucketName:             "artifact_backup_droplet_bucket",
					BucketRegion:           "artifact_backup_droplet_region",
					Blobs:                  []string{dropletsBlob1, dropletsBlob2},
					SrcBackupDirectoryPath: "timestamp/droplets",
				},
				"packages": {
					SameBucketAs: "droplets",
				},
			}
			bucketPairs = map[string]incremental.RestoreBucketPair{
				"droplets": dropletsBucketPair,
				"packages": {
					SameAsBucketID: "droplets",
				},
			}
			artifact.LoadReturns(backups, nil)
		})

		It("does not restore the marked bucket", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(destinationLiveDropletsBucket.CopyBlobFromBucketCallCount()).To(Equal(2))
		})
	})

	Context("When the there is a blob in the artifact that is not in the backup directory", func() {
		BeforeEach(func() {
			artifact.LoadReturns(backups, nil)
			sourceBackupPackagesBucket.ListBlobsReturns(nil, nil)
		})

		It("returns an error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("found blobs in artifact that are not present in backup directory for bucket artifact_backup_package_bucket:"))
			Expect(err.Error()).To(ContainSubstring(packagesBlob1))
			Expect(err.Error()).To(ContainSubstring(packagesBlob2))
			Expect(sourceBackupPackagesBucket.ListBlobsCallCount()).To(Equal(1))
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
			notInArtifactPair := incremental.RestoreBucketPair{
				ConfigLiveBucket:     notInArtifactBucket1,
				ArtifactBackupBucket: notInArtifactBucket2,
			}

			bucketPairs["not-in-artifact"] = notInArtifactPair
		})

		It("returns an error", func() {
			Expect(err).To(MatchError("cannot restore bucket not-in-artifact, not found in backup artifact"))
		})
	})

	Context("When there is a bucket pair that is recorded to have been empty on backup", func() {

		BeforeEach(func() {
			backups = map[string]incremental.Backup{
				"droplets": {
					BucketName:             "artifact_backup_droplet_bucket",
					BucketRegion:           "artifact_backup_droplet_region",
					Blobs:                  []string{dropletsBlob1, dropletsBlob2},
					SrcBackupDirectoryPath: "timestamp/droplets",
				},
				"packages": {
					BucketName:             "artifact_backup_package_bucket",
					BucketRegion:           "artifact_backup_package_region",
					Blobs:                  []string{},
					SrcBackupDirectoryPath: "timestamp/packages",
				},
			}
			artifact.LoadReturns(backups, nil)
		})

		It("does not attempt to restore that pair", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(destinationLiveDropletsBucket.CopyBlobFromBucketCallCount()).To(Equal(2))
			Expect(destinationLivePackagesBucket.CopyBlobFromBucketCallCount()).To(Equal(0))
		})
	})

	Context("When there is a bucket referenced in the artifact that is not in the restore config", func() {
		BeforeEach(func() {
			backups["not-in-restore-config"] = incremental.Backup{
				BucketName:             "whatever",
				BucketRegion:           "whatever",
				Blobs:                  []string{"timestamp/not-in-restore-config/thing"},
				SrcBackupDirectoryPath: "timestamp/not-in-restore-config",
			}
			artifact.LoadReturns(backups, nil)
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
