package unversioned_test

import (
	"fmt"

	"github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/unversioned"
	"github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/unversioned/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Backuper", func() {

	var (
		dropletsBucketPair            *fakes.FakeBucketPair
		buildpacksBucketPair          *fakes.FakeBucketPair
		artifact                      *fakes.FakeArtifact
		fakeClock                     *fakes.FakeClock
		backuper                      unversioned.Backuper
		dropletsBackupBucketAddress   unversioned.BackupBucketAddress
		buildpacksBackupBucketAddress unversioned.BackupBucketAddress
		err                           error
	)

	BeforeEach(func() {
		dropletsBucketPair = new(fakes.FakeBucketPair)
		dropletsBackupBucketAddress = unversioned.BackupBucketAddress{
			BucketName:   "the-droplets-backup-bucket",
			BucketRegion: "the-droplets-backup-region",
			Path:         "time-now-in-seconds/droplets",
		}
		dropletsBucketPair.BackupReturns(dropletsBackupBucketAddress, nil)

		buildpacksBucketPair = new(fakes.FakeBucketPair)
		buildpacksBackupBucketAddress = unversioned.BackupBucketAddress{
			BucketName:   "the-buildpacks-backup-bucket",
			BucketRegion: "the-buildpacks-backup-region",
			Path:         "time-now-in-seconds/buildpacks",
		}
		buildpacksBucketPair.BackupReturns(buildpacksBackupBucketAddress, nil)

		bucketPairs := map[string]unversioned.BucketPair{
			"droplets":   dropletsBucketPair,
			"buildpacks": buildpacksBucketPair,
		}

		artifact = new(fakes.FakeArtifact)

		fakeClock = new(fakes.FakeClock)
		fakeClock.NowReturns("time-now-in-seconds")

		backuper = unversioned.NewBackuper(bucketPairs, artifact, fakeClock)
	})

	JustBeforeEach(func() {
		err = backuper.Run()
	})

	It("copies from the live bucket to the backup bucket", func() {
		Expect(dropletsBucketPair.BackupCallCount()).To(Equal(1))
		Expect(dropletsBucketPair.BackupArgsForCall(0)).To(Equal("time-now-in-seconds/droplets"))
		Expect(buildpacksBucketPair.BackupCallCount()).To(Equal(1))
		Expect(buildpacksBucketPair.BackupArgsForCall(0)).To(Equal("time-now-in-seconds/buildpacks"))
	})

	It("saves the artifact file", func() {
		Expect(artifact.SaveCallCount()).To(Equal(1))
		Expect(artifact.SaveArgsForCall(0)).To(Equal(map[string]unversioned.BackupBucketAddress{
			"droplets":   dropletsBackupBucketAddress,
			"buildpacks": buildpacksBackupBucketAddress,
		}))
	})

	Context("when any of the BucketPairs fails to backup", func() {
		BeforeEach(func() {
			buildpacksBucketPair.BackupReturns(unversioned.BackupBucketAddress{}, fmt.Errorf("BACKUP ERROR"))
		})

		It("exits gracefully", func() {
			By("throwing an error", func() {
				Expect(err).To(MatchError("BACKUP ERROR"))
			})

			By("not saving an artifact", func() {
				Expect(artifact.SaveCallCount()).To(Equal(0))
			})
		})
	})

	Context("When saving the artifact fails", func() {
		BeforeEach(func() {
			artifact.SaveReturns(fmt.Errorf("SAVE ERROR"))
		})

		It("throws an error", func() {
			Expect(err).To(MatchError("SAVE ERROR"))
		})
	})
})
