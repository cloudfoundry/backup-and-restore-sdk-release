package incremental_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/incremental"
	"github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/incremental/fakes"
)

var _ = Describe("BackupStarter", func() {
	It("if finds the last complete backup", func() {
		backupBucket := new(fakes.FakeBucket)
		backupFinder := new(fakes.FakeBackupFinder)
		starter := incremental.BackupStarter{
			BucketPair: incremental.BucketPair{
				BackupBucket: backupBucket,
			},
			BackupFinder: backupFinder,
		}

		err := starter.Run()

		Expect(err).NotTo(HaveOccurred())
		Expect(backupFinder.FindCallCount()).To(Equal(1))
	})

	Context("and finding the last backup fails", func() {
		It("returns an error", func() {
			backupBucket := new(fakes.FakeBucket)
			backupFinder := new(fakes.FakeBackupFinder)
			backupFinder.FindReturns(incremental.BackupDirectory{}, errors.New("fake error"))
			starter := incremental.BackupStarter{
				BucketPair: incremental.BucketPair{
					BackupBucket: backupBucket,
				},
				BackupFinder: backupFinder,
			}

			err := starter.Run()

			Expect(err).To(MatchError(SatisfyAll(
				ContainSubstring("failed to start backup"),
				ContainSubstring("fake error"),
			)))
		})
	})
})
