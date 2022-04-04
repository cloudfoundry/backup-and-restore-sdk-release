package gcs_test

import (
	"errors"

	"github.com/cloudfoundry-incubator/bosh-backup-and-restore/executor"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/backup-and-restore-sdk-release/src/gcs-blobstore-backup-restore"
	"github.com/cloudfoundry/backup-and-restore-sdk-release/src/gcs-blobstore-backup-restore/fakes"
)

var _ = Describe("Backuper", func() {
	var liveBucket *fakes.FakeBucket
	var backupBucket *fakes.FakeBucket
	var backuper gcs.Backuper

	const LiveBucketName = "first-bucket-name"
	const BackupBucketName = "second-bucket-name"
	const bucketId = "bucket-id"
	const blob1 = "file_1_a"
	const blob2 = "file_1_b"

	BeforeEach(func() {
		liveBucket = new(fakes.FakeBucket)
		liveBucket.NameReturns(LiveBucketName)
		backupBucket = new(fakes.FakeBucket)
		backupBucket.NameReturns(BackupBucketName)

		parallelExecutor := executor.NewParallelExecutor()
		parallelExecutor.SetMaxInFlight(200)
		backuper = gcs.NewBackuper(map[string]gcs.BackupToComplete{
			bucketId: {
				BucketPair: gcs.BucketPair{
					LiveBucket:        liveBucket,
					BackupBucket:      backupBucket,
					ExecutionStrategy: parallelExecutor,
				},
				SameAsBucketID: "",
			},
		})
	})
	Context("when there are blobs in the live bucket", func() {

		Context("and it cant list blobs in the live bucket", func() {
			BeforeEach(func() {
				liveBucket.ListBlobsReturns(nil, errors.New("cannot list live blobs"))
			})

			It("it fails with the correct error", func() {
				_, err := backuper.Backup()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("cannot list live blobs"))
			})
		})

		Context("and it cant copy blobs from the live bucket to the backup bucket", func() {
			BeforeEach(func() {
				liveBucket.ListBlobsReturns([]gcs.Blob{
					gcs.NewBlob(blob1),
					gcs.NewBlob(blob2),
				}, nil)
				liveBucket.CopyBlobToBucketReturns(errors.New("cannot copy blob"))
			})

			It("it fails with the correct error", func() {
				_, err := backuper.Backup()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("cannot copy blob"))
			})
		})

		Context("and it successfully runs backup", func() {
			BeforeEach(func() {
				liveBucket.ListBlobsReturns([]gcs.Blob{
					gcs.NewBlob(blob1),
					gcs.NewBlob(blob2),
				}, nil)
			})

			It("tries to copy across the blobs in live bucket to backup bucket", func() {
				_, err := backuper.Backup()
				Expect(err).NotTo(HaveOccurred())

				Expect(liveBucket.CopyBlobToBucketCallCount()).To(Equal(2))
				_, blob1Name, _ := liveBucket.CopyBlobToBucketArgsForCall(0)
				_, blob2Name, _ := liveBucket.CopyBlobToBucketArgsForCall(1)
				argsForCalls := []string{blob1Name, blob2Name}
				Expect(argsForCalls).To(ConsistOf(blob1, blob2))
			})

			It("returns a valid BucketBackup map", func() {
				backupBucketDir, err := backuper.Backup()
				Expect(err).NotTo(HaveOccurred())

				Expect(backupBucketDir[bucketId].BucketName).To(Equal(backupBucket.Name()))
				Expect(backupBucketDir[bucketId].Path).To(MatchRegexp(".*\\d{4}_\\d{2}_\\d{2}_\\d{2}_\\d{2}_\\d{2}/.*"))
			})

			Context("and there is another live bucket which points to the same live bucket", func() {
				const secondBucketId = "another-bucket"

				BeforeEach(func() {
					parallelExecutor := executor.NewParallelExecutor()
					parallelExecutor.SetMaxInFlight(200)
					backuper = gcs.NewBackuper(map[string]gcs.BackupToComplete{
						bucketId: {
							BucketPair: gcs.BucketPair{
								LiveBucket:        liveBucket,
								BackupBucket:      backupBucket,
								ExecutionStrategy: parallelExecutor,
							},
							SameAsBucketID: "",
						},
						secondBucketId: {
							BucketPair: gcs.BucketPair{
								LiveBucket:        liveBucket,
								BackupBucket:      backupBucket,
								ExecutionStrategy: parallelExecutor,
							},
							SameAsBucketID: bucketId,
						},
					})
				})

				It("only copies across the blobs in live bucket to backup bucket once", func() {
					_, err := backuper.Backup()
					Expect(err).NotTo(HaveOccurred())

					Expect(liveBucket.CopyBlobToBucketCallCount()).To(Equal(2))
					_, blob1Name, _ := liveBucket.CopyBlobToBucketArgsForCall(0)
					_, blob2Name, _ := liveBucket.CopyBlobToBucketArgsForCall(1)
					argsForCalls := []string{blob1Name, blob2Name}
					Expect(argsForCalls).To(ConsistOf(blob1, blob2))
				})

				It("returns a valid BucketBackup map", func() {
					backupBucketDir, err := backuper.Backup()
					Expect(err).NotTo(HaveOccurred())

					Expect(backupBucketDir).To(HaveLen(2))

					Expect(backupBucketDir[bucketId].BucketName).To(Equal(backupBucket.Name()))
					Expect(backupBucketDir[bucketId].Path).To(MatchRegexp(".*\\d{4}_\\d{2}_\\d{2}_\\d{2}_\\d{2}_\\d{2}/.*"))
					Expect(backupBucketDir[bucketId].SameBucketAs).To(BeEmpty())

					Expect(backupBucketDir[secondBucketId].BucketName).To(BeEmpty())
					Expect(backupBucketDir[secondBucketId].Path).To(BeEmpty())
					Expect(backupBucketDir[secondBucketId].SameBucketAs).To(Equal(bucketId))
				})
			})
		})
	})
})
