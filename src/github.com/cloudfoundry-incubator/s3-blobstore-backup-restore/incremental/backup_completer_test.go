package incremental_test

import (
	"errors"
	"fmt"

	"github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/incremental"
	"github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/incremental/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("BackupCompleter", func() {
	Context("when there are two blobs to copy", func() {
		var (
			bucket            *fakes.FakeBucket
			dir               incremental.BackupDirectory
			blob1             incremental.BackedUpBlob
			blob2             incremental.BackedUpBlob
			backupsToComplete map[string]incremental.BackupToComplete
		)

		BeforeEach(func() {
			bucket = new(fakes.FakeBucket)
			dir = incremental.BackupDirectory{
				Path:   "timestamp/bucket_id",
				Bucket: bucket,
			}
			blob1 = incremental.BackedUpBlob{
				Path:                "previous_timestamp/bucket_id/f0/fd/blob1/uuid",
				BackupDirectoryPath: "previous_timestamp/bucket_id",
			}
			blob2 = incremental.BackedUpBlob{
				Path:                "previous_timestamp/bucket_id/f0/bucket_id/blob2/uuid",
				BackupDirectoryPath: "previous_timestamp/bucket_id",
			}

			backupsToComplete = map[string]incremental.BackupToComplete{
				"bucket_id": {
					BackupBucket:    bucket,
					BackupDirectory: dir,
					BlobsToCopy:     []incremental.BackedUpBlob{blob1, blob2},
				},
			}
		})

		It("copies the blobs to the backup directory", func() {
			completer := incremental.BackupCompleter{
				BackupsToComplete: backupsToComplete,
			}
			bucket.CopyBlobWithinBucketStub = func(src, dst string) error {
				switch src {
				case "previous_timestamp/bucket_id/f0/fd/blob1/uuid":
					Expect(dst).To(Equal("timestamp/bucket_id/f0/fd/blob1/uuid"))
				case "previous_timestamp/bucket_id/f0/bucket_id/blob2/uuid":
					Expect(dst).To(Equal("timestamp/bucket_id/f0/bucket_id/blob2/uuid"))
				default:
					Fail(fmt.Sprintf("CopyBlobWithinBucket called with unexpected src: %s, dst: %s", src, dst))
				}
				return nil
			}

			err := completer.Run()

			Expect(err).NotTo(HaveOccurred())
			Expect(bucket.CopyBlobWithinBucketCallCount()).To(Equal(2))
		})

		It("marks the backup directory complete", func() {
			completer := incremental.BackupCompleter{
				BackupsToComplete: backupsToComplete,
			}

			err := completer.Run()

			Expect(err).NotTo(HaveOccurred())
			Expect(bucket.UploadBlobCallCount()).To(Equal(1))
		})

		Context("and copy fails for both", func() {
			It("returns an error", func() {
				bucket.CopyBlobWithinBucketReturnsOnCall(0, errors.New("some fake error"))
				bucket.CopyBlobWithinBucketReturnsOnCall(1, errors.New("another fake error"))
				completer := incremental.BackupCompleter{
					BackupsToComplete: backupsToComplete,
				}

				err := completer.Run()

				Expect(err).To(MatchError(SatisfyAll(
					ContainSubstring("failed to complete backup"),
					ContainSubstring("some fake error"),
					ContainSubstring("another fake error"),
				)))
			})
		})

		Context("and a mark complete fails", func() {
			It("returns an error", func() {
				bucket.UploadBlobReturns(errors.New("fake error"))
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

	Context("when there are no blobs to copy", func() {
		It("marks the backup directory complete", func() {
			bucket := new(fakes.FakeBucket)
			dir := incremental.BackupDirectory{
				Bucket: bucket,
			}
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
			Expect(bucket.UploadBlobCallCount()).To(Equal(1))
		})
	})

	Context("when there are no backups to complete", func() {
		It("no-ops", func() {
			backupsToComplete := map[string]incremental.BackupToComplete{}
			completer := incremental.BackupCompleter{
				BackupsToComplete: backupsToComplete,
			}

			err := completer.Run()

			Expect(err).NotTo(HaveOccurred())
		})
	})
})
