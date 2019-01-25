package incremental_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/incremental"
	"github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/incremental/fakes"
)

var _ = Describe("BackupStarter", func() {
	XIt("finds the last complete backup", func() {
		backupBucket := new(fakes.FakeBucket)
		finder := incremental.BackupDirectoryFinder{}
		starter := incremental.BackupStarter{
			BackupsToStart: map[string]incremental.BackupsToStart{
				"bucket_id": {
					BucketPair: incremental.BucketPair{
						BackupBucket: backupBucket,
					},
					BackupDirectoryFinder: finder,
				},
			},
		}

		err := starter.Run()

		Expect(err).NotTo(HaveOccurred())
		//Expect(finder.FindCallCount()).To(Equal(1))
	})

	Context("and finding the last backup fails", func() {
		XIt("returns an error", func() {
			backupBucket := new(fakes.FakeBucket)
			finder := incremental.BackupDirectoryFinder{}
			//finder.FindReturns(incremental.BackupDirectory{}, errors.New("fake error"))
			starter := incremental.BackupStarter{
				BackupsToStart: map[string]incremental.BackupsToStart{
					"bucket_id": {
						BucketPair: incremental.BucketPair{
							BackupBucket: backupBucket,
						},
						BackupDirectoryFinder: finder,
					},
				},
			}

			err := starter.Run()

			Expect(err).To(MatchError(SatisfyAll(
				ContainSubstring("failed to start backup"),
				ContainSubstring("fake error"),
			)))
		})
	})
})
