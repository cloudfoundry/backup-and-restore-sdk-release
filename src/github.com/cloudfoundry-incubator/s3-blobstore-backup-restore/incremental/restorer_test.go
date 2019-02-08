package incremental_test

import (
	"fmt"

	. "github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/incremental"
	"github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/incremental/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Restorer", func() {
	var (
		dropletsBucketPair *fakes.FakeRestoreBucketPair
		packagesBucketPair *fakes.FakeRestoreBucketPair

		bucketPairs map[string]RestoreBucketPair

		artifact *fakes.FakeArtifact

		err error

		restorer Restorer

		bucketBackups map[string]BucketBackup
	)

	BeforeEach(func() {
		dropletsBucketPair = new(fakes.FakeRestoreBucketPair)
		packagesBucketPair = new(fakes.FakeRestoreBucketPair)

		artifact = new(fakes.FakeArtifact)
		bucketBackups = map[string]BucketBackup{
			"droplets": {
				BucketName:          "artifact_backup_droplet_bucket",
				BucketRegion:        "artifact_backup_droplet_region",
				Blobs:               []string{"my_key", "another_key"},
				BackupDirectoryPath: "timestamp/droplets",
			},
			"packages": {
				BucketName:          "artifact_backup_package_bucket",
				BucketRegion:        "artifact_backup_package_region",
				Blobs:               []string{"my_key", "another_key"},
				BackupDirectoryPath: "timestamp2/packages",
			},
		}
		artifact.LoadReturns(bucketBackups, nil)

		bucketPairs = map[string]RestoreBucketPair{
			"droplets": dropletsBucketPair,
			"packages": packagesBucketPair,
		}

		restorer = NewRestorer(bucketPairs, artifact)
	})

	JustBeforeEach(func() {
		err = restorer.Run()
	})

	Context("When the artifact is valid and copying works", func() {
		BeforeEach(func() {
			dropletsBucketPair.RestoreReturns(nil)
			packagesBucketPair.RestoreReturns(nil)

		})

		It("restores all the bucket pairs", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(dropletsBucketPair.RestoreCallCount()).To(Equal(1))
			Expect(dropletsBucketPair.RestoreArgsForCall(0)).To(Equal(bucketBackups["droplets"]))
			Expect(packagesBucketPair.RestoreCallCount()).To(Equal(1))
			Expect(packagesBucketPair.RestoreArgsForCall(0)).To(Equal(bucketBackups["packages"]))
		})
	})

	Context("When a bucket cannot be restored", func() {
		BeforeEach(func() {
			dropletsBucketPair.RestoreReturns(fmt.Errorf("restore error"))
		})

		It("returns an error", func() {
			Expect(err).To(MatchError("restore error"))
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
		var notInArtifactPair *fakes.FakeRestoreBucketPair

		BeforeEach(func() {
			notInArtifactPair = new(fakes.FakeRestoreBucketPair)

			bucketPairs = map[string]RestoreBucketPair{
				"droplets":        dropletsBucketPair,
				"packages":        packagesBucketPair,
				"not-in-artifact": notInArtifactPair,
			}
			restorer = NewRestorer(bucketPairs, artifact)
		})

		It("returns an error", func() {
			Expect(err).To(MatchError("cannot restore bucket not-in-artifact, not found in backup artifact"))
		})
	})

	Context("When there is a bucket pair that is recorded to have been empty on backup", func() {

		BeforeEach(func() {
			bucketBackups = map[string]BucketBackup{
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
					BackupDirectoryPath: "timestamp2/packages",
				},
			}
			artifact.LoadReturns(bucketBackups, nil)
		})

		It("does not attempt to restore that pair", func() {
			Expect(dropletsBucketPair.RestoreCallCount()).To(Equal(1))
			Expect(packagesBucketPair.RestoreCallCount()).To(Equal(0))
		})
	})

	Context("When there is a bucket referenced in the artifact that is not in the restore config", func() {
		BeforeEach(func() {
			bucketBackups = map[string]BucketBackup{
				"droplets": {
					BucketName:          "artifact_backup_droplet_bucket",
					BucketRegion:        "artifact_backup_droplet_region",
					Blobs:               []string{"my_key", "another_key"},
					BackupDirectoryPath: "timestamp/droplets",
				},
				"packages": {
					BucketName:          "artifact_backup_package_bucket",
					BucketRegion:        "artifact_backup_package_region",
					Blobs:               []string{"my_key", "another_key"},
					BackupDirectoryPath: "timestamp2/packages",
				},
				"not-in-restore-config": {
					BucketName:          "whatever",
					BucketRegion:        "whatever",
					Blobs:               []string{"thing"},
					BackupDirectoryPath: "timestamp2/not-in-restore-config",
				},
			}
			artifact.LoadReturns(bucketBackups, nil)
		})

		It("returns an error", func() {
			Expect(err).To(MatchError("bucket not-in-restore-config is not mentioned in the restore config" +
				" but is present in the artifact"))
		})
	})
})
