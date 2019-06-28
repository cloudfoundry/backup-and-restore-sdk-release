package incremental_test

import (
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/backup-and-restore-sdk/s3-blobstore-backup-restore/incremental"
	"github.com/cloudfoundry-incubator/backup-and-restore-sdk/s3-blobstore-backup-restore/incremental/fakes"
)

var _ = Describe("BackupStarter", func() {
	var (
		err                   error
		clock                 *fakes.FakeClock
		backupBucket          *fakes.FakeBucket
		liveBucket            *fakes.FakeBucket
		liveBlob1             *fakes.FakeBlob
		liveBlob2             *fakes.FakeBlob
		liveBlob3             *fakes.FakeBlob
		backupCompleteBlob    *fakes.FakeBlob
		artifact              *fakes.FakeArtifact
		existingBlobsArtifact *fakes.FakeArtifact
		backupDirectoryFinder *fakes.FakeBackupDirectoryFinder
		starter               incremental.BackupStarter
	)

	BeforeEach(func() {
		clock = new(fakes.FakeClock)
		clock.NowReturns("2000_01_02_03_04_05")
		backupBucket = new(fakes.FakeBucket)
		backupBucket.NameReturns("backup-bucket")
		backupBucket.RegionReturns("us-east-1")
		liveBucket = new(fakes.FakeBucket)
		artifact = new(fakes.FakeArtifact)
		existingBlobsArtifact = new(fakes.FakeArtifact)

		liveBlob1 = new(fakes.FakeBlob)
		liveBlob1.PathReturns("f0/fd/blob1/uuid")
		liveBlob2 = new(fakes.FakeBlob)
		liveBlob2.PathReturns("f0/fd/blob2/uuid")
		liveBlob3 = new(fakes.FakeBlob)
		liveBlob3.PathReturns("f0/fd/blob3/uuid")
		backupCompleteBlob = new(fakes.FakeBlob)
		backupCompleteBlob.PathReturns("backup_complete")

		backupDirectoryFinder = new(fakes.FakeBackupDirectoryFinder)

		starter = incremental.NewBackupStarter(
			map[string]incremental.BackupToStart{
				"bucket_id": {
					BucketPair: incremental.BackupBucketPair{
						ConfigBackupBucket: backupBucket,
						ConfigLiveBucket:   liveBucket,
					},
					BackupDirectoryFinder: backupDirectoryFinder,
				},
			},
			clock,
			artifact,
			existingBlobsArtifact,
		)
	})

	JustBeforeEach(func() {
		err = starter.Run()
	})

	Context("when there are no previous complete backups", func() {
		BeforeEach(func() {
			backupDirectoryFinder.ListBlobsReturns(nil, nil)

			liveBlobs := []incremental.Blob{liveBlob1, liveBlob2, liveBlob3, backupCompleteBlob}
			liveBucket.ListBlobsReturns(liveBlobs, nil)
		})

		It("copies all the live blobs to the new backup directory", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(backupDirectoryFinder.ListBlobsCallCount()).To(Equal(1))
			Expect(liveBucket.ListBlobsCallCount()).To(Equal(1))
			Expect(liveBucket.ListBlobsArgsForCall(0)).To(Equal(""))

			Expect(backupBucket.CopyBlobFromBucketCallCount()).To(Equal(3))

			bucket0, src0, dst0 := backupBucket.CopyBlobFromBucketArgsForCall(0)
			bucket1, src1, dst1 := backupBucket.CopyBlobFromBucketArgsForCall(1)
			bucket2, src2, dst2 := backupBucket.CopyBlobFromBucketArgsForCall(2)
			Expect([]incremental.Bucket{bucket0, bucket1, bucket2}).To(ConsistOf(liveBucket, liveBucket, liveBucket))
			Expect([][]string{{src0, dst0}, {src1, dst1}, {src2, dst2}}).To(ConsistOf(
				[]string{
					"f0/fd/blob1/uuid",
					"2000_01_02_03_04_05/bucket_id/f0/fd/blob1/uuid",
				},
				[]string{
					"f0/fd/blob2/uuid",
					"2000_01_02_03_04_05/bucket_id/f0/fd/blob2/uuid",
				},
				[]string{
					"f0/fd/blob3/uuid",
					"2000_01_02_03_04_05/bucket_id/f0/fd/blob3/uuid",
				},
			))
		})

		It("writes the backup artifact", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(artifact.WriteCallCount()).To(Equal(1))
			Expect(artifact.WriteArgsForCall(0)).To(Equal(map[string]incremental.Backup{
				"bucket_id": {
					BucketName: "backup-bucket",
					Blobs: []string{
						"2000_01_02_03_04_05/bucket_id/f0/fd/blob1/uuid",
						"2000_01_02_03_04_05/bucket_id/f0/fd/blob2/uuid",
						"2000_01_02_03_04_05/bucket_id/f0/fd/blob3/uuid",
					},
					SrcBackupDirectoryPath: "2000_01_02_03_04_05/bucket_id",
					BucketRegion:           "us-east-1",
				},
			}))
		})

		It("writes an empty existing blobs artifact", func() {
			Expect(err).NotTo(HaveOccurred())

			Expect(existingBlobsArtifact.WriteCallCount()).To(Equal(1))
			Expect(existingBlobsArtifact.WriteArgsForCall(0)).To(Equal(map[string]incremental.Backup{
				"bucket_id": {
					DstBackupDirectoryPath: "2000_01_02_03_04_05/bucket_id",
				},
			}))
		})

		Context("and listing the blobs from the live bucket fails", func() {
			BeforeEach(func() {
				liveBucket.ListBlobsReturns(nil, fmt.Errorf("mayhem"))
			})

			It("returns an error", func() {
				Expect(backupDirectoryFinder.ListBlobsCallCount()).To(Equal(1))
				Expect(liveBucket.ListBlobsCallCount()).To(Equal(1))
				Expect(err).To(MatchError(SatisfyAll(
					ContainSubstring("failed to start backup"),
					ContainSubstring("mayhem"),
				)))
			})
		})

		Context("and a backup to start is marked the same as another", func() {
			BeforeEach(func() {
				starter = incremental.NewBackupStarter(
					map[string]incremental.BackupToStart{
						"bucket_id": {
							BucketPair: incremental.BackupBucketPair{
								ConfigBackupBucket: backupBucket,
								ConfigLiveBucket:   liveBucket,
							},
							BackupDirectoryFinder: backupDirectoryFinder,
						},
						"marked_bucket_id": {
							SameAsBucketID: "bucket_id",
						},
					},
					clock,
					artifact,
					existingBlobsArtifact,
				)
			})

			It("backs up the bucket only once", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(backupBucket.CopyBlobFromBucketCallCount()).To(Equal(3))
			})

			It("writes the backup artifact", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(artifact.WriteCallCount()).To(Equal(1))
				Expect(artifact.WriteArgsForCall(0)).To(Equal(map[string]incremental.Backup{
					"bucket_id": {
						BucketName: "backup-bucket",
						Blobs: []string{
							"2000_01_02_03_04_05/bucket_id/f0/fd/blob1/uuid",
							"2000_01_02_03_04_05/bucket_id/f0/fd/blob2/uuid",
							"2000_01_02_03_04_05/bucket_id/f0/fd/blob3/uuid",
						},
						SrcBackupDirectoryPath: "2000_01_02_03_04_05/bucket_id",
						BucketRegion:           "us-east-1",
					},
					"marked_bucket_id": {
						SameBucketAs: "bucket_id",
					},
				}))
			})

			It("writes an empty existing blobs artifact", func() {
				Expect(err).NotTo(HaveOccurred())

				Expect(existingBlobsArtifact.WriteCallCount()).To(Equal(1))
				Expect(existingBlobsArtifact.WriteArgsForCall(0)).To(Equal(map[string]incremental.Backup{
					"bucket_id": {
						DstBackupDirectoryPath: "2000_01_02_03_04_05/bucket_id",
					},
					"marked_bucket_id": {
						SameBucketAs: "bucket_id",
					},
				}))
			})
		})
	})

	Context("when there is a previous complete backup", func() {
		BeforeEach(func() {
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
		})

		It("copies only the new live blobs to the new backup directory", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(backupDirectoryFinder.ListBlobsCallCount()).To(Equal(1))
			Expect(liveBucket.ListBlobsCallCount()).To(Equal(1))

			Expect(backupBucket.CopyBlobFromBucketCallCount()).To(Equal(1))

			actualBucket, liveBlobPath, blobDst := backupBucket.CopyBlobFromBucketArgsForCall(0)
			Expect(actualBucket).To(Equal(liveBucket))
			Expect(liveBlobPath).To(Equal("f0/fd/blob3/uuid"))
			Expect(blobDst).To(Equal("2000_01_02_03_04_05/bucket_id/f0/fd/blob3/uuid"))
		})

		It("writes the previously-backed up blobs to the backup directory", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(backupDirectoryFinder.ListBlobsCallCount()).To(Equal(1))
			Expect(liveBucket.ListBlobsCallCount()).To(Equal(1))

			Expect(backupBucket.CopyBlobFromBucketCallCount()).To(Equal(1))
			Expect(artifact.WriteCallCount()).To(Equal(1))
			Expect(artifact.WriteArgsForCall(0)).To(Equal(map[string]incremental.Backup{
				"bucket_id": {
					BucketName: "backup-bucket",
					Blobs: []string{
						"2000_01_02_03_04_05/bucket_id/f0/fd/blob1/uuid",
						"2000_01_02_03_04_05/bucket_id/f0/fd/blob2/uuid",
						"2000_01_02_03_04_05/bucket_id/f0/fd/blob3/uuid",
					},
					SrcBackupDirectoryPath: "2000_01_02_03_04_05/bucket_id",
					BucketRegion:           "us-east-1",
				},
			}))

			Expect(existingBlobsArtifact.WriteCallCount()).To(Equal(1))
			Expect(existingBlobsArtifact.WriteArgsForCall(0)).To(Equal(map[string]incremental.Backup{
				"bucket_id": {
					Blobs: []string{
						"2000_01_01_01_01_01/bucket_id/f0/fd/blob1/uuid",
						"2000_01_01_01_01_01/bucket_id/f0/fd/blob2/uuid",
					},
					SrcBackupDirectoryPath: "2000_01_01_01_01_01/bucket_id",
					DstBackupDirectoryPath: "2000_01_02_03_04_05/bucket_id",
				},
			}))
		})

		It("writes the backup artifact", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(backupDirectoryFinder.ListBlobsCallCount()).To(Equal(1))
			Expect(liveBucket.ListBlobsCallCount()).To(Equal(1))
			Expect(backupBucket.CopyBlobFromBucketCallCount()).To(Equal(1))
			Expect(artifact.WriteCallCount()).To(Equal(1))
			Expect(artifact.WriteArgsForCall(0)).To(Equal(map[string]incremental.Backup{
				"bucket_id": {
					BucketName: "backup-bucket",
					Blobs: []string{
						"2000_01_02_03_04_05/bucket_id/f0/fd/blob1/uuid",
						"2000_01_02_03_04_05/bucket_id/f0/fd/blob2/uuid",
						"2000_01_02_03_04_05/bucket_id/f0/fd/blob3/uuid",
					},
					SrcBackupDirectoryPath: "2000_01_02_03_04_05/bucket_id",
					BucketRegion:           "us-east-1",
				},
			}))
		})

		Context("and the copying a blob from live bucket to backup bucket fails", func() {
			BeforeEach(func() {
				backupDirectoryFinder.ListBlobsReturns(nil, nil)

				liveBucket.ListBlobsReturns([]incremental.Blob{liveBlob1, liveBlob2}, nil)
				backupBucket.CopyBlobFromBucketReturnsOnCall(0, errors.New("some copy error"))
				backupBucket.CopyBlobFromBucketReturnsOnCall(1, errors.New("another copy error"))
			})

			It("returns an error", func() {
				Expect(err).To(MatchError(SatisfyAll(
					ContainSubstring("failed to copy blobs during backup"),
					ContainSubstring("some copy error"),
					ContainSubstring("another copy error"),
				)))
			})
		})

		Context("and when writing the previously backed-up blobs fails", func() {
			BeforeEach(func() {
				existingBlobsArtifact.WriteReturns(errors.New("fake error"))
			})

			It("returns an error", func() {
				Expect(err).To(MatchError(SatisfyAll(
					ContainSubstring("failed to write existing blobs artifact"),
					ContainSubstring("fake error"),
				)))
			})
		})

		Context("and finding the last backup fails", func() {
			BeforeEach(func() {
				backupDirectoryFinder.ListBlobsReturns(nil, errors.New("fake error"))
			})

			It("returns an error", func() {
				Expect(err).To(MatchError(SatisfyAll(
					ContainSubstring("failed to start backup"),
					ContainSubstring("fake error"),
				)))
			})
		})

		Context("and when the backupArtifact fails to write", func() {
			BeforeEach(func() {
				artifact.WriteReturns(errors.New("backupArtifact no"))
			})

			It("returns an error", func() {
				Expect(err).To(MatchError(SatisfyAll(
					ContainSubstring("failed to write backup artifact"),
					ContainSubstring("backupArtifact no"),
				)))
			})
		})
	})
})
