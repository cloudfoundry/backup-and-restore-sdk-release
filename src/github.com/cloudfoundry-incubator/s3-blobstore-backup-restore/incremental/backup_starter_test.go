package incremental_test

import (
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/incremental"
	"github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/incremental/fakes"
)

var _ = Describe("BackupStarter", func() {
	var (
		clock                 *fakes.FakeClock
		backupBucket          *fakes.FakeBucket
		liveBucket            *fakes.FakeBucket
		liveBlob1             *fakes.FakeBlob
		liveBlob2             *fakes.FakeBlob
		liveBlob3             *fakes.FakeBlob
		artifact              *fakes.FakeArtifact
		backupDirectoryFinder *fakes.FakeBackupDirectoryFinder
	)

	BeforeEach(func() {
		clock = new(fakes.FakeClock)
		clock.NowReturns("2000_01_02_03_04_05")
		backupBucket = new(fakes.FakeBucket)
		backupBucket.NameReturns("backup-bucket")
		liveBucket = new(fakes.FakeBucket)
		artifact = new(fakes.FakeArtifact)

		liveBlob1 = new(fakes.FakeBlob)
		liveBlob1.PathReturns("f0/fd/blob1/uuid")
		liveBlob2 = new(fakes.FakeBlob)
		liveBlob2.PathReturns("f0/fd/blob2/uuid")
		liveBlob3 = new(fakes.FakeBlob)
		liveBlob3.PathReturns("f0/fd/blob3/uuid")

		backupDirectoryFinder = new(fakes.FakeBackupDirectoryFinder)
	})

	It("finds the blobs from the last complete backup", func() {
		starter := incremental.NewBackupStarter(
			map[string]incremental.BackupsToStart{
				"bucket_id": {
					BucketPair: incremental.BucketPair{
						BackupBucket: backupBucket,
						LiveBucket:   liveBucket,
					},
					BackupDirectoryFinder: backupDirectoryFinder,
				},
			},
			clock,
			artifact,
		)

		err := starter.Run()

		Expect(err).NotTo(HaveOccurred())
		Expect(backupDirectoryFinder.ListBlobsCallCount()).To(Equal(1))
	})

	It("finds the blobs in the live bucket", func() {
		backupDirectoryFinder.ListBlobsReturns(nil, nil)
		liveBucket.ListBlobsReturns(nil, nil)

		starter := incremental.NewBackupStarter(
			map[string]incremental.BackupsToStart{
				"bucket_id": {
					BucketPair: incremental.BucketPair{
						LiveBucket:   liveBucket,
						BackupBucket: backupBucket,
					},
					BackupDirectoryFinder: backupDirectoryFinder,
				},
			},
			clock,
			artifact,
		)

		err := starter.Run()

		Expect(err).NotTo(HaveOccurred())
		Expect(backupDirectoryFinder.ListBlobsCallCount()).To(Equal(1))
		Expect(liveBucket.ListBlobsCallCount()).To(Equal(1))
	})

	Context("when there are no previous complete backups", func() {
		It("copies all the live blobs to the new backup directory", func() {
			backupDirectoryFinder.ListBlobsReturns(nil, nil)

			liveBlobs := []incremental.Blob{liveBlob1, liveBlob2, liveBlob3}
			liveBucket.ListBlobsReturns(liveBlobs, nil)

			starter := incremental.NewBackupStarter(
				map[string]incremental.BackupsToStart{
					"bucket_id": {
						BucketPair: incremental.BucketPair{
							LiveBucket:   liveBucket,
							BackupBucket: backupBucket,
						},
						BackupDirectoryFinder: backupDirectoryFinder,
					},
				},
				clock,
				artifact,
			)

			err := starter.Run()

			Expect(err).NotTo(HaveOccurred())
			Expect(backupDirectoryFinder.ListBlobsCallCount()).To(Equal(1))
			Expect(liveBucket.ListBlobsCallCount()).To(Equal(1))

			Expect(liveBucket.CopyBlobToBucketCallCount()).To(Equal(3))
			backupBucketArg, liveBlobPath, blobDst := liveBucket.CopyBlobToBucketArgsForCall(0)
			Expect(backupBucketArg).To(Equal(backupBucket))
			Expect(liveBlobPath).To(Equal("f0/fd/blob1/uuid"))
			Expect(blobDst).To(Equal("2000_01_02_03_04_05/bucket_id/f0/fd/blob1/uuid"))

			backupBucketArg, liveBlobPath2, blobDst2 := liveBucket.CopyBlobToBucketArgsForCall(1)
			Expect(backupBucketArg).To(Equal(backupBucket))
			Expect(liveBlobPath2).To(Equal("f0/fd/blob2/uuid"))
			Expect(blobDst2).To(Equal("2000_01_02_03_04_05/bucket_id/f0/fd/blob2/uuid"))

			backupBucketArg, liveBlobPath3, blobDst3 := liveBucket.CopyBlobToBucketArgsForCall(2)
			Expect(backupBucketArg).To(Equal(backupBucket))
			Expect(liveBlobPath3).To(Equal("f0/fd/blob3/uuid"))
			Expect(blobDst3).To(Equal("2000_01_02_03_04_05/bucket_id/f0/fd/blob3/uuid"))
		})
	})

	Context("when there is a previous complete backup", func() {
		It("copies only the new live blobs to the new backup directory", func() {
			backedUpBlob1 := incremental.BackedUpBlob{
				Path:                "2000_01_01_01_01_01/bucket_id/f0/fd/blob1/uuid",
				BackupDirectoryPath: "2000_01_01_01_01_01/bucket_id",
			}
			backedUpBlob2 := incremental.BackedUpBlob{
				Path:                "2000_01_01_01_01_01/bucket_id/f0/fd/blob2/uuid",
				BackupDirectoryPath: "2000_01_01_01_01_01/bucket_id",
			}

			backupDirectoryFinder.ListBlobsReturns([]incremental.BackedUpBlob{backedUpBlob1, backedUpBlob2}, nil)

			liveBucket.ListBlobsReturns([]incremental.Blob{liveBlob1, liveBlob2, liveBlob3}, nil)

			starter := incremental.NewBackupStarter(
				map[string]incremental.BackupsToStart{
					"bucket_id": {
						BucketPair: incremental.BucketPair{
							LiveBucket:   liveBucket,
							BackupBucket: backupBucket,
						},
						BackupDirectoryFinder: backupDirectoryFinder,
					},
				},
				clock,
				artifact,
			)

			err := starter.Run()

			Expect(err).NotTo(HaveOccurred())
			Expect(backupDirectoryFinder.ListBlobsCallCount()).To(Equal(1))
			Expect(liveBucket.ListBlobsCallCount()).To(Equal(1))

			Expect(liveBucket.CopyBlobToBucketCallCount()).To(Equal(1))

			backupBucketArg, liveBlobPath, blobDst := liveBucket.CopyBlobToBucketArgsForCall(0)
			Expect(backupBucketArg).To(Equal(backupBucket))
			Expect(liveBlobPath).To(Equal("f0/fd/blob3/uuid"))
			Expect(blobDst).To(Equal("2000_01_02_03_04_05/bucket_id/f0/fd/blob3/uuid"))
		})

		Context("and the copying a blob from live bucket to backup bucket fails", func() {
			It("returns an error", func() {
				backupDirectoryFinder.ListBlobsReturns(nil, nil)

				liveBucket.ListBlobsReturns([]incremental.Blob{liveBlob1}, nil)
				liveBucket.CopyBlobToBucketReturns(errors.New("oups"))

				starter := incremental.NewBackupStarter(
					map[string]incremental.BackupsToStart{
						"bucket_id": {
							BucketPair: incremental.BucketPair{
								LiveBucket:   liveBucket,
								BackupBucket: backupBucket,
							},
							BackupDirectoryFinder: backupDirectoryFinder,
						},
					},
					clock,
					artifact,
				)

				err := starter.Run()

				Expect(err).To(MatchError(SatisfyAll(
					ContainSubstring("failed to copy blobs during backup"),
					ContainSubstring("oups"),
				)))
			})
		})
	})

	Context("and finding the last backup fails", func() {
		It("returns an error", func() {
			backupDirectoryFinder.ListBlobsReturns(nil, errors.New("fake error"))
			starter := incremental.NewBackupStarter(
				map[string]incremental.BackupsToStart{
					"bucket_id": {
						BucketPair: incremental.BucketPair{
							BackupBucket: backupBucket,
						},
						BackupDirectoryFinder: backupDirectoryFinder,
					},
				},
				clock,
				artifact,
			)

			err := starter.Run()

			Expect(err).To(MatchError(SatisfyAll(
				ContainSubstring("failed to start backup"),
				ContainSubstring("fake error"),
			)))
		})
	})

	Context("and listing the blobs from the live bucket", func() {
		It("returns an error", func() {
			backupDirectoryFinder.ListBlobsReturns(nil, nil)
			liveBucket.ListBlobsReturns(nil, fmt.Errorf("mayhem"))

			starter := incremental.NewBackupStarter(
				map[string]incremental.BackupsToStart{
					"bucket_id": {
						BucketPair: incremental.BucketPair{
							LiveBucket:   liveBucket,
							BackupBucket: backupBucket,
						},
						BackupDirectoryFinder: backupDirectoryFinder,
					},
				},
				clock,
				artifact,
			)

			err := starter.Run()

			Expect(backupDirectoryFinder.ListBlobsCallCount()).To(Equal(1))
			Expect(liveBucket.ListBlobsCallCount()).To(Equal(1))
			Expect(err).To(MatchError(SatisfyAll(
				ContainSubstring("failed to start backup"),
				ContainSubstring("mayhem"),
			)))
		})
	})

	It("writes the backup artifact", func() {
		backupDirectoryFinder.ListBlobsReturns(nil, nil)

		liveBlobs := []incremental.Blob{liveBlob1, liveBlob2}
		liveBucket.ListBlobsReturns(liveBlobs, nil)

		starter := incremental.NewBackupStarter(
			map[string]incremental.BackupsToStart{
				"bucket_id": {
					BucketPair: incremental.BucketPair{
						LiveBucket:   liveBucket,
						BackupBucket: backupBucket,
					},
					BackupDirectoryFinder: backupDirectoryFinder,
				},
			},
			clock,
			artifact,
		)

		err := starter.Run()

		Expect(err).NotTo(HaveOccurred())
		Expect(backupDirectoryFinder.ListBlobsCallCount()).To(Equal(1))
		Expect(liveBucket.ListBlobsCallCount()).To(Equal(1))
		Expect(liveBucket.CopyBlobToBucketCallCount()).To(Equal(2))
		Expect(artifact.WriteCallCount()).To(Equal(1))
		Expect(artifact.WriteArgsForCall(0)).To(Equal(map[string]incremental.BucketBackup{
			"bucket_id": {
				BucketName: "backup-bucket",
				Blobs: []string{
					"2000_01_02_03_04_05/bucket_id/f0/fd/blob1/uuid",
					"2000_01_02_03_04_05/bucket_id/f0/fd/blob2/uuid",
				},
				BackupDirectoryPath: "2000_01_02_03_04_05/bucket_id",
			},
		}))
	})

	Context("when the artifact fails to write", func() {
		It("returns an error", func() {
			backupDirectoryFinder.ListBlobsReturns(nil, nil)

			liveBlobs := []incremental.Blob{liveBlob1, liveBlob2}
			liveBucket.ListBlobsReturns(liveBlobs, nil)

			artifact.WriteReturns(errors.New("artifact no"))
			starter := incremental.NewBackupStarter(
				map[string]incremental.BackupsToStart{
					"bucket_id": {
						BucketPair: incremental.BucketPair{
							LiveBucket:   liveBucket,
							BackupBucket: backupBucket,
						},
						BackupDirectoryFinder: backupDirectoryFinder,
					},
				},
				clock,
				artifact,
			)

			err := starter.Run()

			Expect(err).To(MatchError(SatisfyAll(
				ContainSubstring("failed to write artifact"),
				ContainSubstring("artifact no"),
			)))

		})
	})
})
