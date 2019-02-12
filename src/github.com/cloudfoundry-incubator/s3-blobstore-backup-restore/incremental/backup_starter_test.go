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
		err                   error
		clock                 *fakes.FakeClock
		backupBucket          *fakes.FakeBucket
		liveBucket            *fakes.FakeBucket
		liveBlob1             *fakes.FakeBlob
		liveBlob2             *fakes.FakeBlob
		liveBlob3             *fakes.FakeBlob
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

		backupDirectoryFinder = new(fakes.FakeBackupDirectoryFinder)

		starter = incremental.NewBackupStarter(
			map[string]incremental.BackupToStart{
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
			existingBlobsArtifact,
		)
	})

	JustBeforeEach(func() {
		err = starter.Run()
	})

	Context("when there are no previous complete backups", func() {
		BeforeEach(func() {
			backupDirectoryFinder.ListBlobsReturns(nil, nil)

			liveBlobs := []incremental.Blob{liveBlob1, liveBlob2, liveBlob3}
			liveBucket.ListBlobsReturns(liveBlobs, nil)

			backupBucket.CopyBlobFromBucketStub = func(bucket incremental.Bucket, src, dst string) error {
				Expect(bucket).To(Equal(liveBucket))
				switch src {
				case "f0/fd/blob1/uuid":
					Expect(dst).To(Equal("2000_01_02_03_04_05/bucket_id/f0/fd/blob1/uuid"))
				case "f0/fd/blob2/uuid":
					Expect(dst).To(Equal("2000_01_02_03_04_05/bucket_id/f0/fd/blob2/uuid"))
				case "f0/fd/blob3/uuid":
					Expect(dst).To(Equal("2000_01_02_03_04_05/bucket_id/f0/fd/blob3/uuid"))
				default:
					Fail(fmt.Sprintf("CopyBlobFromBucket unexpected src: %s, with dst: %s, bucket: %s", src, dst, bucket))
				}
				return nil
			}
		})

		It("copies all the live blobs to the new backup directory", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(backupDirectoryFinder.ListBlobsCallCount()).To(Equal(1))
			Expect(liveBucket.ListBlobsCallCount()).To(Equal(1))
			Expect(liveBucket.ListBlobsArgsForCall(0)).To(Equal(""))

			Expect(backupBucket.CopyBlobFromBucketCallCount()).To(Equal(3))
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
			Expect(artifact.WriteArgsForCall(0)).To(Equal(map[string]incremental.BucketBackup{
				"bucket_id": {
					BucketName: "backup-bucket",
					Blobs: []string{
						"2000_01_02_03_04_05/bucket_id/f0/fd/blob1/uuid",
						"2000_01_02_03_04_05/bucket_id/f0/fd/blob2/uuid",
						"2000_01_02_03_04_05/bucket_id/f0/fd/blob3/uuid",
					},
					BackupDirectoryPath: "2000_01_02_03_04_05/bucket_id",
					BucketRegion:        "us-east-1",
				},
			}))

			Expect(existingBlobsArtifact.WriteCallCount()).To(Equal(1))
			Expect(existingBlobsArtifact.WriteArgsForCall(0)).To(Equal(map[string]incremental.BucketBackup{
				"bucket_id": {
					BucketName: "backup-bucket",
					Blobs: []string{
						"2000_01_01_01_01_01/bucket_id/f0/fd/blob1/uuid",
						"2000_01_01_01_01_01/bucket_id/f0/fd/blob2/uuid",
					},
					BackupDirectoryPath: "2000_01_01_01_01_01/bucket_id",
					BucketRegion:        "us-east-1",
				},
			}))
		})

		It("writes the backup backupArtifact", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(backupDirectoryFinder.ListBlobsCallCount()).To(Equal(1))
			Expect(liveBucket.ListBlobsCallCount()).To(Equal(1))
			Expect(backupBucket.CopyBlobFromBucketCallCount()).To(Equal(1))
			Expect(artifact.WriteCallCount()).To(Equal(1))
			Expect(artifact.WriteArgsForCall(0)).To(Equal(map[string]incremental.BucketBackup{
				"bucket_id": {
					BucketName: "backup-bucket",
					Blobs: []string{
						"2000_01_02_03_04_05/bucket_id/f0/fd/blob1/uuid",
						"2000_01_02_03_04_05/bucket_id/f0/fd/blob2/uuid",
						"2000_01_02_03_04_05/bucket_id/f0/fd/blob3/uuid",
					},
					BackupDirectoryPath: "2000_01_02_03_04_05/bucket_id",
					BucketRegion:        "us-east-1",
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
					ContainSubstring("failed to write existing blobs backupArtifact"),
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
					ContainSubstring("failed to write backupArtifact"),
					ContainSubstring("backupArtifact no"),
				)))

			})
		})
	})
})
