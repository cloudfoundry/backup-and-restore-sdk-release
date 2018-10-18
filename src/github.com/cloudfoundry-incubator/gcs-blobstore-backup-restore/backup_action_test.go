package gcs_test

import (
	"fmt"

	. "github.com/cloudfoundry-incubator/gcs-blobstore-backup-restore"
	"github.com/cloudfoundry-incubator/gcs-blobstore-backup-restore/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("BackupAction", func() {
	var backuper *fakes.FakeBackuper
	var artifact *fakes.FakeBackupArtifact
	var backupAction BackupAction
	var err error

	BeforeEach(func() {
		backuper = new(fakes.FakeBackuper)
		artifact = new(fakes.FakeBackupArtifact)
		backupAction = NewBackupAction()
	})

	Context("when everything goes well", func() {
		It("succeeds", func() {
			By("running backup steps", func() {
				err = backupAction.Run(backuper, artifact)
				Expect(err).NotTo(HaveOccurred())

				Expect(backuper.CreateLiveBucketSnapshotCallCount()).To(Equal(1))
				Expect(backuper.TransferBlobsToBackupBucketCallCount()).To(Equal(1))
				Expect(backuper.CopyBlobsWithinBackupBucketCallCount()).To(Equal(1))
			})

			By("running cleanup", func() {
				Expect(backuper.CleanupLiveBucketsCallCount()).To(Equal(1))
			})

			By("generating backup artifact", func() {
				Expect(artifact.WriteCallCount()).To(Equal(1))
			})
		})
	})

	Context("when CreateLiveBucketSnapshot fails", func() {
		It("fails with the correct error", func() {
			backuper.CreateLiveBucketSnapshotReturns(fmt.Errorf("I failed to create a live bucket snapshot"))
			err = backupAction.Run(backuper, artifact)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("I failed to create a live bucket snapshot"))

			Expect(backuper.CreateLiveBucketSnapshotCallCount()).To(Equal(1))
			Expect(backuper.TransferBlobsToBackupBucketCallCount()).To(Equal(0))
			Expect(backuper.CopyBlobsWithinBackupBucketCallCount()).To(Equal(0))
			Expect(artifact.WriteCallCount()).To(Equal(0))
		})
	})

	Context("when TransferBlobsToBackupBucket fails", func() {
		It("fails with the correct error", func() {
			backuper.TransferBlobsToBackupBucketReturns(nil, fmt.Errorf("I failed to transfer blobs"))
			err = backupAction.Run(backuper, artifact)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("I failed to transfer blobs"))

			Expect(backuper.CreateLiveBucketSnapshotCallCount()).To(Equal(1))
			Expect(backuper.TransferBlobsToBackupBucketCallCount()).To(Equal(1))
			Expect(backuper.CopyBlobsWithinBackupBucketCallCount()).To(Equal(0))
			Expect(artifact.WriteCallCount()).To(Equal(0))
		})
	})

	Context("when CopyBlobsWithinBackupBucket fails", func() {
		It("fails with the correct error", func() {
			backuper.CopyBlobsWithinBackupBucketReturns(fmt.Errorf("I failed to copy blobs"))
			err = backupAction.Run(backuper, artifact)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("I failed to copy blobs"))

			Expect(backuper.CreateLiveBucketSnapshotCallCount()).To(Equal(1))
			Expect(backuper.TransferBlobsToBackupBucketCallCount()).To(Equal(1))
			Expect(backuper.CopyBlobsWithinBackupBucketCallCount()).To(Equal(1))
			Expect(artifact.WriteCallCount()).To(Equal(0))
		})
	})

	Context("when CleanupLiveBuckets fails", func() {
		It("fails with the correct error", func() {
			backuper.CleanupLiveBucketsReturns(fmt.Errorf("I failed to clean up"))
			err = backupAction.Run(backuper, artifact)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("I failed to clean up"))

			Expect(backuper.CreateLiveBucketSnapshotCallCount()).To(Equal(1))
			Expect(backuper.TransferBlobsToBackupBucketCallCount()).To(Equal(1))
			Expect(backuper.CopyBlobsWithinBackupBucketCallCount()).To(Equal(1))
			Expect(backuper.CleanupLiveBucketsCallCount()).To(Equal(1))

			Expect(artifact.WriteCallCount()).To(Equal(0))
		})
	})

	Context("when artifact fails to write", func() {
		It("fails with the correct error", func() {
			artifact.WriteReturns(fmt.Errorf("I failed to write"))
			err = backupAction.Run(backuper, artifact)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("I failed to write"))

			Expect(backuper.CreateLiveBucketSnapshotCallCount()).To(Equal(1))
			Expect(backuper.TransferBlobsToBackupBucketCallCount()).To(Equal(1))
			Expect(backuper.CopyBlobsWithinBackupBucketCallCount()).To(Equal(1))
			Expect(backuper.CleanupLiveBucketsCallCount()).To(Equal(1))
			Expect(artifact.WriteCallCount()).To(Equal(1))
		})
	})

})
