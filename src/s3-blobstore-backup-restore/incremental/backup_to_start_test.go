package incremental_test

import (
	"s3-blobstore-backup-restore/incremental"
	"s3-blobstore-backup-restore/incremental/fakes"

	. "github.com/onsi/ginkgo/v2"

	. "github.com/onsi/gomega"
)

var _ = Describe("MarkSameBackupsToStart", func() {
	It("marks backups to start that are the same as another bucket ID", func() {
		bucket1 := new(fakes.FakeBucket)
		bucket1.NameReturns("live-bucket-1")
		bucket3 := new(fakes.FakeBucket)
		bucket3.NameReturns("live-bucket-3")

		backupsToStart := map[string]incremental.BackupToStart{
			"bucket2": {
				BucketPair: incremental.BackupBucketPair{
					ConfigLiveBucket: bucket1,
				},
			},
			"bucket3": {
				BucketPair: incremental.BackupBucketPair{
					ConfigLiveBucket: bucket3,
				},
			},
			"bucket4": {
				BucketPair: incremental.BackupBucketPair{
					ConfigLiveBucket: bucket1,
				},
			},
			"bucket1": {
				BucketPair: incremental.BackupBucketPair{
					ConfigLiveBucket: bucket1,
				},
			},
		}

		markedBackups := incremental.MarkSameBackupsToStart(backupsToStart)

		Expect(markedBackups).To(Equal(map[string]incremental.BackupToStart{
			"bucket1": {
				BucketPair: incremental.BackupBucketPair{
					ConfigLiveBucket: bucket1,
				},
			},
			"bucket2": {
				SameAsBucketID: "bucket1",
			},
			"bucket3": {
				BucketPair: incremental.BackupBucketPair{
					ConfigLiveBucket: bucket3,
				},
			},
			"bucket4": {
				SameAsBucketID: "bucket1",
			},
		}))
	})
})
