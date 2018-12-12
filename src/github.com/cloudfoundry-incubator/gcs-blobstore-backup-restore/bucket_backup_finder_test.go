package gcs_test

import (
	"errors"
	"fmt"

	"github.com/cloudfoundry-incubator/gcs-blobstore-backup-restore"
	"github.com/cloudfoundry-incubator/gcs-blobstore-backup-restore/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("BucketBackupFinder", func() {
	const (
		bucketID = "bucket-id"

		firstBackupDir  = "first-backup-dir"
		secondBackupDir = "second-backup-dir"
		thirdBackupDir  = "third-backup-dir"

		firstBucketBackupDir  = firstBackupDir + "/" + bucketID
		secondBucketBackupDir = secondBackupDir + "/" + bucketID
		thirdBucketBackupDir  = thirdBackupDir + "/" + bucketID
	)

	var (
		backupBucket *fakes.FakeBucket
		backupFinder gcs.BucketBackupFinder
	)

	Describe("ListBlobs", func() {
		BeforeEach(func() {
			backupBucket = new(fakes.FakeBucket)

			backupBucket.ListBackupsReturns([]string{firstBackupDir}, nil)
			backupBucket.IsBackupCompleteReturns(true, nil)

			backupFinder = gcs.NewLastBucketBackupFinder(bucketID, backupBucket)
		})

		Context("when there is a complete backup", func() {
			It("lists all backup blobs", func() {
				someBlob := gcs.NewBlob(fmt.Sprintf("%s/some-blob-name", firstBucketBackupDir))
				anotherBlob := gcs.NewBlob(fmt.Sprintf("%s/another-blob-name", firstBucketBackupDir))
				lastBackupBlobs := []gcs.Blob{someBlob, anotherBlob}
				backupBucket.ListBlobsReturns(lastBackupBlobs, nil)

				blobs, err := backupFinder.ListBlobs()

				Expect(err).NotTo(HaveOccurred())
				By("calling ListBackups", func() {
					Expect(backupBucket.ListBackupsCallCount()).To(Equal(1))
				})

				By("checking if the last directory is complete", func() {
					Expect(backupBucket.IsBackupCompleteCallCount()).To(Equal(1))
					Expect(backupBucket.IsBackupCompleteArgsForCall(0)).To(Equal(firstBucketBackupDir))
				})

				By("listing all blobs from the last directory", func() {
					Expect(backupBucket.ListBlobsCallCount()).To(Equal(1))
					Expect(backupBucket.ListBlobsArgsForCall(0)).To(Equal(firstBackupDir))
				})

				Expect(blobs).To(Equal(map[string]gcs.Blob{
					"some-blob-name":    someBlob,
					"another-blob-name": anotherBlob,
				}))
			})
		})

		Context("when there is no complete backup", func() {
			It("returns an empty blobs map", func() {
				backupBucket.ListBackupsReturns([]string{firstBackupDir, secondBackupDir, thirdBackupDir}, nil)
				backupBucket.IsBackupCompleteReturnsOnCall(0, false, nil)
				backupBucket.IsBackupCompleteReturnsOnCall(1, false, nil)
				backupBucket.IsBackupCompleteReturnsOnCall(2, false, nil)

				backupBlobs, err := backupFinder.ListBlobs()

				Expect(err).NotTo(HaveOccurred())
				Expect(backupBucket.IsBackupCompleteCallCount()).To(Equal(3))
				Expect(backupBucket.IsBackupCompleteArgsForCall(0)).To(Equal(thirdBucketBackupDir))
				Expect(backupBucket.IsBackupCompleteArgsForCall(1)).To(Equal(secondBucketBackupDir))
				Expect(backupBucket.IsBackupCompleteArgsForCall(2)).To(Equal(firstBucketBackupDir))
				Expect(backupBucket.ListBlobsCallCount()).To(BeZero())
				Expect(backupBlobs).To(BeEmpty())
			})
		})

		Context("when there are multiple backups and only one is complete", func() {
			It("lists all the complete backup blobs", func() {
				backupBucket.ListBackupsReturns([]string{firstBackupDir, secondBackupDir, thirdBackupDir}, nil)
				backupBucket.IsBackupCompleteReturnsOnCall(0, false, nil)
				backupBucket.IsBackupCompleteReturnsOnCall(1, true, nil)

				blob := gcs.NewBlob(fmt.Sprintf("%s/some-blob-name", secondBucketBackupDir))
				lastBackupBlobs := []gcs.Blob{blob}
				backupBucket.ListBlobsReturns(lastBackupBlobs, nil)

				blobs, err := backupFinder.ListBlobs()

				Expect(err).ToNot(HaveOccurred())
				Expect(backupBucket.IsBackupCompleteArgsForCall(0)).To(Equal(thirdBucketBackupDir))
				Expect(backupBucket.IsBackupCompleteArgsForCall(1)).To(Equal(secondBucketBackupDir))
				Expect(backupBucket.IsBackupCompleteCallCount()).To(Equal(2))
				Expect(backupBucket.ListBlobsCallCount()).To(Equal(1))
				Expect(backupBucket.ListBlobsArgsForCall(0)).To(Equal(secondBackupDir))
				Expect(blobs).To(Equal(map[string]gcs.Blob{
					"some-blob-name": blob,
				}))
			})
		})

		Context("when there are no directories in the backup bucket", func() {
			It("returns an empty blobs map", func() {
				backupBucket.ListBackupsReturns(nil, nil)

				blobs, err := backupFinder.ListBlobs()

				Expect(err).NotTo(HaveOccurred())
				Expect(blobs).To(BeEmpty())
			})
		})

		It("returns an error when listing directories fails", func() {
			backupBucket.ListBackupsReturns(nil, errors.New("failed"))

			_, err := backupFinder.ListBlobs()

			Expect(err).To(MatchError(ContainSubstring("failed listing last backup blobs")))
			Expect(backupBucket.IsBackupCompleteCallCount()).To(BeZero())
		})

		It("returns an error when checking is complete fails", func() {
			backupBucket.IsBackupCompleteReturns(false, errors.New("failed"))

			_, err := backupFinder.ListBlobs()

			Expect(err).To(MatchError(ContainSubstring("failed listing last backup blobs")))
			Expect(backupBucket.ListBlobsCallCount()).To(BeZero())
		})

		It("returns an error when listing blobs fails", func() {
			backupBucket.ListBlobsReturns(nil, errors.New("failed to list"))

			_, err := backupFinder.ListBlobs()

			Expect(err).To(MatchError(ContainSubstring("failed listing last backup blobs")))
		})
	})
})
