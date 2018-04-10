package unversioned_test

import (
	"fmt"

	"github.com/cloudfoundry-incubator/blobstore-backup-restore/unversioned"
	"github.com/cloudfoundry-incubator/blobstore-backup-restore/unversioned/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Restorer", func() {
	var (
		dropletsBucketPair *fakes.FakeBucketPair
		packagesBucketPair *fakes.FakeBucketPair

		bucketPairs map[string]unversioned.BucketPair

		artifact *fakes.FakeArtifact

		err error

		restorer unversioned.Restorer
	)

	BeforeEach(func() {
		dropletsBucketPair = new(fakes.FakeBucketPair)
		packagesBucketPair = new(fakes.FakeBucketPair)

		artifact = new(fakes.FakeArtifact)
		artifact.LoadReturns(map[string]unversioned.BackupBucketAddress{
			"droplets": {
				BucketName:   "artifact_backup_droplet_bucket",
				BucketRegion: "artifact_backup_droplet_region",
				Path:         "timestamp/droplets",
			},
			"packages": {
				BucketName:   "artifact_backup_package_bucket",
				BucketRegion: "artifact_backup_package_region",
				Path:         "timestamp2/packages",
			},
		}, nil)

		bucketPairs = map[string]unversioned.BucketPair{
			"droplets": dropletsBucketPair,
			"packages": packagesBucketPair,
		}

		restorer = unversioned.NewRestorer(bucketPairs, artifact)
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
			Expect(dropletsBucketPair.RestoreArgsForCall(0)).To(Equal("timestamp/droplets"))
			Expect(packagesBucketPair.RestoreCallCount()).To(Equal(1))
			Expect(packagesBucketPair.RestoreArgsForCall(0)).To(Equal("timestamp2/packages"))
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
		var notInArtifactPair *fakes.FakeBucketPair

		BeforeEach(func() {
			notInArtifactPair = new(fakes.FakeBucketPair)

			bucketPairs = map[string]unversioned.BucketPair{
				"droplets":        dropletsBucketPair,
				"packages":        packagesBucketPair,
				"not-in-artifact": notInArtifactPair,
			}
			restorer = unversioned.NewRestorer(bucketPairs, artifact)
		})

		It("returns an error", func() {
			Expect(err).To(MatchError("cannot restore bucket not-in-artifact, not found in backup artifact"))
		})
	})

	Context("When there is a bucket pair that is recorded to have been empty on backup", func() {

		BeforeEach(func() {

			artifact.LoadReturns(map[string]unversioned.BackupBucketAddress{
				"droplets": {
					BucketName:   "artifact_backup_droplet_bucket",
					BucketRegion: "artifact_backup_droplet_region",
					Path:         "timestamp/droplets",
				},
				"packages": {
					BucketName:   "artifact_backup_package_bucket",
					BucketRegion: "artifact_backup_package_region",
					Path:         "timestamp2/packages",
					EmptyBackup:  true,
				},
			}, nil)
		})

		It("does not attempt to restore that pair", func() {
			Expect(dropletsBucketPair.RestoreCallCount()).To(Equal(1))
			Expect(packagesBucketPair.RestoreCallCount()).To(Equal(0))
		})
	})

	Context("When there is a bucket referenced in the artifact that is not in the restore config", func() {
		BeforeEach(func() {
			artifact.LoadReturns(map[string]unversioned.BackupBucketAddress{
				"droplets": {
					BucketName:   "artifact_backup_droplet_bucket",
					BucketRegion: "artifact_backup_droplet_region",
					Path:         "timestamp/droplets",
				},
				"packages": {
					BucketName:   "artifact_backup_package_bucket",
					BucketRegion: "artifact_backup_package_region",
					Path:         "timestamp2/packages",
				},
				"not-in-restore-config": {
					BucketName:   "whatever",
					BucketRegion: "whatever",
					Path:         "timestamp2/not-in-restore-config",
				},
			}, nil)

		})

		It("returns an error", func() {
			Expect(err).To(MatchError("bucket not-in-restore-config is not mentioned in the restore config" +
				" but is present in the artifact"))
		})
	})
})
