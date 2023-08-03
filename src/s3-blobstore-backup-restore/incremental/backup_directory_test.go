package incremental_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"

	"s3-blobstore-backup-restore/incremental"
	"s3-blobstore-backup-restore/incremental/fakes"

	. "github.com/onsi/gomega"
)

var _ = Describe("BackupDirectory", func() {
	It("marks complete", func() {
		bucket := new(fakes.FakeBucket)
		dir := incremental.BackupDirectory{
			Path:   "timestamp/bucket_id",
			Bucket: bucket,
		}

		err := dir.MarkComplete()

		Expect(err).NotTo(HaveOccurred())
		Expect(bucket.UploadBlobCallCount()).To(Equal(1))
		path, contents := bucket.UploadBlobArgsForCall(0)
		Expect(path).To(Equal("timestamp/bucket_id/backup_complete"))
		Expect(contents).To(Equal(""))
	})

	Context("and uploading backup complete blob fails", func() {
		It("returns an error", func() {
			bucket := new(fakes.FakeBucket)
			bucket.UploadBlobReturns(errors.New("fake error"))
			dir := incremental.BackupDirectory{
				Path:   "timestamp/bucket_id",
				Bucket: bucket,
			}

			err := dir.MarkComplete()

			Expect(err).To(MatchError(SatisfyAll(
				ContainSubstring("failed marking backup directory 'timestamp/bucket_id' as complete"),
				ContainSubstring("fake error"),
			)))
		})
	})

	It("knows it is complete", func() {
		bucket := new(fakes.FakeBucket)
		bucket.HasBlobReturns(true, nil)
		dir := incremental.BackupDirectory{
			Path:   "timestamp/bucket_id",
			Bucket: bucket,
		}

		isComplete, err := dir.IsComplete()

		Expect(err).NotTo(HaveOccurred())
		Expect(isComplete).To(BeTrue())
		Expect(bucket.HasBlobCallCount()).To(Equal(1))
		path := bucket.HasBlobArgsForCall(0)
		Expect(path).To(Equal("timestamp/bucket_id/backup_complete"))
	})

	It("knows it is NOT complete", func() {
		bucket := new(fakes.FakeBucket)
		bucket.HasBlobReturns(false, nil)
		dir := incremental.BackupDirectory{
			Path:   "timestamp/bucket_id",
			Bucket: bucket,
		}

		isComplete, err := dir.IsComplete()

		Expect(err).NotTo(HaveOccurred())
		Expect(isComplete).To(BeFalse())
	})

	Context("and checking if the blob exists fails", func() {
		It("returns an error", func() {
			bucket := new(fakes.FakeBucket)
			bucket.HasBlobReturns(false, errors.New("fake error"))
			dir := incremental.BackupDirectory{
				Path:   "timestamp/bucket_id",
				Bucket: bucket,
			}

			_, err := dir.IsComplete()

			Expect(err).To(MatchError(SatisfyAll(
				ContainSubstring("failed checking if backup directory 'timestamp/bucket_id' is complete"),
				ContainSubstring("fake error"),
			)))
		})
	})

	It("lists blobs in the backup directory", func() {
		bucket := new(fakes.FakeBucket)
		blob1 := new(fakes.FakeBlob)
		blob1.PathReturns("timestamp/bucket_id/fd/f0/blob1/uuid")
		blob2 := new(fakes.FakeBlob)
		blob2.PathReturns("timestamp/bucket_id/fd/f0/blob2/uuid")
		bucket.ListBlobsReturns([]incremental.Blob{
			blob1,
			blob2,
		}, nil)
		dir := incremental.BackupDirectory{
			Path:   "timestamp/bucket_id",
			Bucket: bucket,
		}

		blobs, err := dir.ListBlobs()

		Expect(err).NotTo(HaveOccurred())
		Expect(blobs).To(ConsistOf(
			incremental.BackedUpBlob{
				Path:                "timestamp/bucket_id/fd/f0/blob1/uuid",
				BackupDirectoryPath: "timestamp/bucket_id",
			},
			incremental.BackedUpBlob{
				Path:                "timestamp/bucket_id/fd/f0/blob2/uuid",
				BackupDirectoryPath: "timestamp/bucket_id",
			},
		))
		Expect(bucket.ListBlobsCallCount()).To(Equal(1))
		Expect(bucket.ListBlobsArgsForCall(0)).To(Equal("timestamp/bucket_id"))
	})

	Context("and listing blobs fails", func() {
		It("returns an error", func() {
			bucket := new(fakes.FakeBucket)
			bucket.ListBlobsReturns(nil, errors.New("fake error"))
			dir := incremental.BackupDirectory{
				Path:   "timestamp/bucket_id",
				Bucket: bucket,
			}

			_, err := dir.ListBlobs()

			Expect(err).To(MatchError(SatisfyAll(
				ContainSubstring("failed listing blobs in backup directory 'timestamp/bucket_id'"),
				ContainSubstring("fake error"),
			)))
		})
	})
})
