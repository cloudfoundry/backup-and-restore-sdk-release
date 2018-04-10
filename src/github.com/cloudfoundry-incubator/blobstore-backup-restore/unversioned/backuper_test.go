package unversioned_test

import (
	"fmt"

	"github.com/cloudfoundry-incubator/blobstore-backup-restore/unversioned"
	"github.com/cloudfoundry-incubator/blobstore-backup-restore/unversioned/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Backuper", func() {

	var (
		dropletBucketPair   *fakes.FakeBucketPair
		artifact            *fakes.FakeArtifact
		fakeClock           *fakes.FakeClock
		backuper            unversioned.Backuper
		backupBucketAddress unversioned.BackupBucketAddress
		err                 error
	)

	BeforeEach(func() {
		dropletBucketPair = new(fakes.FakeBucketPair)
		backupBucketAddress = unversioned.BackupBucketAddress{
			BucketName:   "the-backup-bucket",
			BucketRegion: "the-backup-region",
			Path:         "time-now-in-seconds/droplets",
		}
		dropletBucketPair.BackupReturns(backupBucketAddress, nil)
		artifact = new(fakes.FakeArtifact)
		bucketPairs := map[string]unversioned.BucketPair{
			"droplets": dropletBucketPair,
		}
		fakeClock = new(fakes.FakeClock)
		fakeClock.NowReturns("time-now-in-seconds")
		backuper = unversioned.NewBackuper(bucketPairs, artifact, fakeClock)
	})

	JustBeforeEach(func() {
		err = backuper.Run()
	})

	It("copies from the live bucket to the backup bucket", func() {
		Expect(dropletBucketPair.BackupCallCount()).To(Equal(1))
		Expect(dropletBucketPair.BackupArgsForCall(0)).To(Equal("time-now-in-seconds/droplets"))
	})

	It("saves the artifact file", func() {
		Expect(artifact.SaveCallCount()).To(Equal(1))
		Expect(artifact.SaveArgsForCall(0)).To(Equal(map[string]unversioned.BackupBucketAddress{
			"droplets": backupBucketAddress,
		}))
	})

	Context("When the Backup call fails", func() {
		BeforeEach(func() {
			dropletBucketPair.BackupReturns(unversioned.BackupBucketAddress{}, fmt.Errorf("BACKUP ERROR"))
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
