package incremental_test

import (
	"errors"

	"github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/incremental"
	"github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/incremental/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("BackupCompleter", func() {
	Context("when there is two blobs to copy", func() {
		var (
			bucket            *fakes.FakeBucket
			dir               *fakes.FakeBackupDirectory
			blob1             *fakes.FakeBlob
			blob2             *fakes.FakeBlob
			backupsToComplete map[string]incremental.BackupToComplete
		)

		BeforeEach(func() {
			bucket = new(fakes.FakeBucket)
			dir = new(fakes.FakeBackupDirectory)
			dir.PathReturns("timestamp")
			blob1 = new(fakes.FakeBlob)
			blob2 = new(fakes.FakeBlob)
			blob1.NameReturns("previous_timestamp/bucket_id/f0/fd/blob1/uuid")
			blob2.NameReturns("previous_timestamp/bucket_id/f0/bucket_id/blob2/uuid")

			backupsToComplete = map[string]incremental.BackupToComplete{
				"bucket_id": {
					BackupBucket:    bucket,
					BackupDirectory: dir,
					BlobsToCopy:     []incremental.Blob{blob1, blob2},
				},
			}
		})

		It("copies the blobs to the backup directory", func() {
			completer := incremental.BackupCompleter{
				BackupsToComplete: backupsToComplete,
			}

			err := completer.Run()

			Expect(err).NotTo(HaveOccurred())
			Expect(bucket.CopyBlobWithinBucketCallCount()).To(Equal(2))
			src, dst := bucket.CopyBlobWithinBucketArgsForCall(0)
			Expect(src).To(Equal("previous_timestamp/bucket_id/f0/fd/blob1/uuid"))
			Expect(dst).To(Equal("timestamp/bucket_id/f0/fd/blob1/uuid"))
			src, dst = bucket.CopyBlobWithinBucketArgsForCall(1)
			Expect(src).To(Equal("previous_timestamp/bucket_id/f0/bucket_id/blob2/uuid"))
			Expect(dst).To(Equal("timestamp/bucket_id/f0/bucket_id/blob2/uuid"))
		})

		It("marks the backup directory complete", func() {
			completer := incremental.BackupCompleter{
				BackupsToComplete: backupsToComplete,
			}

			err := completer.Run()

			Expect(err).NotTo(HaveOccurred())
			Expect(dir.MarkCompleteCallCount()).To(Equal(1))
		})

		Context("and a copy fails", func() {
			It("returns an error", func() {
				bucket.CopyBlobWithinBucketReturns(errors.New("fake error"))
				completer := incremental.BackupCompleter{
					BackupsToComplete: backupsToComplete,
				}

				err := completer.Run()

				Expect(err).To(MatchError(SatisfyAll(
					ContainSubstring("failed to complete backup"),
					ContainSubstring("fake error"),
				)))
			})
		})

		Context("and a mark complete fails", func() {
			It("returns an error", func() {
				dir.MarkCompleteReturns(errors.New("fake error"))
				completer := incremental.BackupCompleter{
					BackupsToComplete: backupsToComplete,
				}

				err := completer.Run()

				Expect(err).To(MatchError(SatisfyAll(
					ContainSubstring("failed to complete backup"),
					ContainSubstring("fake error"),
				)))
			})
		})
	})

	Context("when there are no backup directories", func() {
		It("no-ops", func() {
			bucket := new(fakes.FakeBucket)
			dir := new(fakes.FakeBackupDirectory)
			backupsToComplete := map[string]incremental.BackupToComplete{
				"bucket_id": {
					BackupBucket:    bucket,
					BackupDirectory: dir,
					BlobsToCopy:     nil,
				},
			}
			completer := incremental.BackupCompleter{
				BackupsToComplete: backupsToComplete,
			}

			err := completer.Run()

			Expect(err).NotTo(HaveOccurred())
			Expect(bucket.CopyBlobWithinBucketCallCount()).To(BeZero())
		})
	})
})
