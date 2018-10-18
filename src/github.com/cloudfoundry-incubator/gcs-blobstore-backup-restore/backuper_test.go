package gcs_test

import (
	"encoding/json"
	"errors"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/gcs-blobstore-backup-restore"
	"github.com/cloudfoundry-incubator/gcs-blobstore-backup-restore/fakes"
)

var _ = Describe("GCSBackuper", func() {
	Describe("CreateLiveBucketSnapshot", func() {
		var bucket *fakes.FakeBucket
		var backupBucket *fakes.FakeBucket

		var backuper gcs.GCSBackuper

		const firstBucketName = "first-bucket-name"
		const secondBucketName = "second-bucket-name"

		BeforeEach(func() {
			bucket = new(fakes.FakeBucket)
			bucket.NameReturns(firstBucketName)
			backupBucket = new(fakes.FakeBucket)
			backupBucket.NameReturns(secondBucketName)

			backuper = gcs.NewBackuper(map[string]gcs.BucketPair{
				"first": {
					Bucket:       bucket,
					BackupBucket: backupBucket,
				},
			})
		})

		Context("when there is no previous backup artifact", func() {
			Context("and there is a single bucket to be backed up", func() {
				var (
					blob1 string
					blob2 string
				)

				BeforeEach(func() {
					blob1 = "file_1_a"
					blob2 = "file_1_b"
					bucket.ListBlobsReturns([]gcs.Blob{
						{Name: blob1},
						{Name: blob2},
					}, nil)

					bucket.CopyBlobWithinBucketReturns(nil)
					bucket.CreateFileReturns(nil)
				})

				It("creates a snapshot directory with a copy of the live bucket and an empty common blobs file", func() {

					err := backuper.CreateLiveBucketSnapshot()
					Expect(err).NotTo(HaveOccurred())

					Expect(bucket.CopyBlobWithinBucketCallCount()).To(Equal(2))
					blob, path := bucket.CopyBlobWithinBucketArgsForCall(0)
					Expect(blob).To(Equal(blob1))
					Expect(path).To(Equal(fmt.Sprintf("temporary-backup-artifact/%s", blob1)))

					blob, path = bucket.CopyBlobWithinBucketArgsForCall(1)
					Expect(blob).To(Equal(blob2))
					Expect(path).To(Equal(fmt.Sprintf("temporary-backup-artifact/%s", blob2)))

					Expect(bucket.CreateFileCallCount()).To(Equal(1))
					file, contents := bucket.CreateFileArgsForCall(0)
					Expect(file).To(Equal("temporary-backup-artifact/common_blobs.json"))
					Expect(contents).To(Equal([]byte("null")))
				})
			})
		})

		Context("when there is a previous backup artifact", func() {
			var blob1, blob2 string
			BeforeEach(func() {
				blob1 = "file_1_a"
				blob2 = "file_1_b"
				bucket.ListBlobsReturns([]gcs.Blob{
					{Name: blob1},
					{Name: blob2},
				}, nil)

				backupBucket.ListLastBackupBlobsReturns([]gcs.Blob{
					{Name: "1970_01_01_00_00_00/droplets/" + blob1},
				}, nil)
			})

			It("creates a snapshot directory with a delta between the live and backup buckets and a common blobs file", func() {
				err := backuper.CreateLiveBucketSnapshot()
				Expect(err).NotTo(HaveOccurred())

				bucket.CopyBlobWithinBucketReturns(nil)
				Expect(bucket.CopyBlobWithinBucketCallCount()).To(Equal(1))
				blob, path := bucket.CopyBlobWithinBucketArgsForCall(0)
				Expect(blob).To(Equal(blob2))
				Expect(path).To(Equal(fmt.Sprintf("temporary-backup-artifact/%s", blob2)))

				Expect(bucket.CreateFileCallCount()).To(Equal(1))
				file, contents := bucket.CreateFileArgsForCall(0)
				Expect(file).To(Equal("temporary-backup-artifact/common_blobs.json"))
				j, _ := json.Marshal([]gcs.Blob{{Name: "1970_01_01_00_00_00/droplets/" + blob1}})
				Expect(contents).To(Equal(j))
			})

			Context("when the delta is empty", func() {
				BeforeEach(func() {
					blob1 = "file_1_a"
					bucket.ListBlobsReturns([]gcs.Blob{
						{Name: blob1},
					}, nil)

					backupBucket.ListLastBackupBlobsReturns([]gcs.Blob{
						{Name: "1970_01_01_00_00_00/droplets/" + blob1},
					}, nil)
				})

				It("creates an empty snapshot directory in the live bucket", func() {
					err := backuper.CreateLiveBucketSnapshot()
					Expect(err).NotTo(HaveOccurred())

					Expect(bucket.CopyBlobWithinBucketCallCount()).To(Equal(0))
					Expect(bucket.CreateFileCallCount()).To(Equal(1))
					filePath, _ := bucket.CreateFileArgsForCall(0)
					Expect(filePath).To(Equal("temporary-backup-artifact/common_blobs.json"))
				})
			})
		})

		Context("when list blobs fails", func() {
			It("returns an error", func() {
				bucket.ListBlobsReturns(nil, errors.New("ifailed"))
				err := backuper.CreateLiveBucketSnapshot()
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("ifailed"))
			})
		})

		Context("when copy blob fails", func() {
			It("returns an error", func() {
				blob1 := "file_1_a"
				bucket.ListBlobsReturns([]gcs.Blob{
					{Name: blob1},
				}, nil)

				bucket.CopyBlobWithinBucketReturns(errors.New("oopsifailed"))
				err := backuper.CreateLiveBucketSnapshot()
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("oopsifailed"))
			})
		})
	})

	Describe("TransferBlobsToBackupBucket", func() {
		var bucket *fakes.FakeBucket
		var backupBucket *fakes.FakeBucket
		var bucketPairID = "droplets"

		var backuper gcs.GCSBackuper

		const firstBucketName = "first-bucket-name"

		BeforeEach(func() {
			bucket = new(fakes.FakeBucket)
			bucket.NameReturns(firstBucketName)

			backupBucket = new(fakes.FakeBucket)
			backupBucket.NameReturns(firstBucketName)

			backuper = gcs.NewBackuper(map[string]gcs.BucketPair{
				bucketPairID: {
					Bucket:       bucket,
					BackupBucket: backupBucket,
				},
			})
		})

		Context("when there is no previous backup artifact", func() {
			Context("and there is a single bucket to be backed up", func() {
				var (
					blob1, blob2 string
				)

				BeforeEach(func() {
					blob1 = "file_1_a"
					blob2 = "temporary-backup-artifact/file_1_b"
					bucket.ListBlobsReturns([]gcs.Blob{
						{Name: blob1},
						{Name: blob2},
					}, nil)

					bucket.CopyBlobBetweenBucketsReturns(nil)
				})
				It("transfers the blobs from the live bucket to the backup bucket", func() {
					_, err := backuper.TransferBlobsToBackupBucket()
					Expect(err).NotTo(HaveOccurred())

					Expect(bucket.CopyBlobBetweenBucketsCallCount()).To(Equal(1))
					dstBucket, blob, path := bucket.CopyBlobBetweenBucketsArgsForCall(0)
					Expect(dstBucket.Name()).To(Equal(backupBucket.Name()))
					Expect(blob).To(Equal(blob2))
					Expect(path).To(MatchRegexp("\\d{4}_\\d{2}_\\d{2}_\\d{2}_\\d{2}_\\d{2}/%s/file_1_b", bucketPairID))
				})

				It("returns a map of the backup buckets and paths", func() {
					backupBuckets, err := backuper.TransferBlobsToBackupBucket()
					Expect(err).NotTo(HaveOccurred())

					Expect(backupBuckets).To(HaveLen(1))
					Expect(backupBuckets[bucketPairID].BucketName).To(Equal(bucket.Name()))
					Expect(backupBuckets[bucketPairID].Path).To(MatchRegexp("\\d{4}_\\d{2}_\\d{2}_\\d{2}_\\d{2}_\\d{2}/%s", bucketPairID))
				})
			})
		})

		Context("when list blobs fails", func() {
			It("returns an error", func() {
				bucket.ListBlobsReturns(nil, errors.New("ifailed"))
				_, err := backuper.TransferBlobsToBackupBucket()
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("ifailed"))
			})
		})

		Context("when copy blob fails", func() {
			It("returns an error", func() {
				blob1 := "temporary-backup-artifact/file_1_a"
				bucket.ListBlobsReturns([]gcs.Blob{
					{Name: blob1},
				}, nil)

				bucket.CopyBlobBetweenBucketsReturns(errors.New("oopsifailed"))
				_, err := backuper.TransferBlobsToBackupBucket()
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("oopsifailed"))
			})
		})
	})

	Describe("CleanupLiveBuckets", func() {
		var bucket *fakes.FakeBucket
		var bucketPairID = "droplets"

		var backuper gcs.GCSBackuper

		const firstBucketName = "first-bucket-name"

		BeforeEach(func() {
			bucket = new(fakes.FakeBucket)
			bucket.NameReturns(firstBucketName)
			bucket.ListBlobsReturns([]gcs.Blob{{Name: "temporary-backup-artifact/common_blobs.json"}}, nil)

			backuper = gcs.NewBackuper(map[string]gcs.BucketPair{
				bucketPairID: {
					Bucket: bucket,
				},
			})
		})

		It("Deletes live bucket backup artifact", func() {
			err := backuper.CleanupLiveBuckets()
			Expect(err).NotTo(HaveOccurred())
			Expect(bucket.DeleteBlobCallCount()).To(Equal(1))
			Expect(bucket.DeleteBlobArgsForCall(0)).To(Equal("temporary-backup-artifact/common_blobs.json"))
		})

		Context("when it fails to delete a blob", func() {
			BeforeEach(func() {
				bucket.DeleteBlobReturns(fmt.Errorf("I failed to delete"))
			})

			It("reports an error", func() {
				err := backuper.CleanupLiveBuckets()
				Expect(err).To(MatchError("I failed to delete"))
			})
		})

		Context("when it fails to list the blobs", func() {
			BeforeEach(func() {
				bucket.ListBlobsReturns(nil, fmt.Errorf("I failed to list blobs"))
			})

			It("reports an error", func() {
				err := backuper.CleanupLiveBuckets()
				Expect(err).To(MatchError("I failed to list blobs"))
			})
		})
	})

	Describe("CopyBlobsWithinBackupBucket", func() {
		var bucket *fakes.FakeBucket
		var backupBucket *fakes.FakeBucket
		var bucketPairID = "droplets"

		var backuper gcs.GCSBackuper

		const firstBucketName = "first-bucket-name"

		BeforeEach(func() {
			bucket = new(fakes.FakeBucket)
			bucket.NameReturns(firstBucketName)

			backupBucket = new(fakes.FakeBucket)
			backupBucket.NameReturns(firstBucketName)

			backuper = gcs.NewBackuper(map[string]gcs.BucketPair{
				bucketPairID: {
					Bucket:       bucket,
					BackupBucket: backupBucket,
				},
			})
		})

		Context("when all of the blobs are common", func() {
			var blob1 string
			backupBucketAddresses := make(map[string]gcs.BackupBucketAddress)

			BeforeEach(func() {
				blob1 = "file1"
				bucket.ListBlobsReturns([]gcs.Blob{
					{Name: blob1},
					{Name: "temporary_backup_artifact/common_blobs.json"},
				}, nil)
				backupBucket.ListLastBackupBlobsReturns([]gcs.Blob{
					{Name: "1970_01_01_00_00_00/droplets/" + blob1},
				}, nil)
				backupBucket.CopyBlobBetweenBucketsReturns(nil)
				backupBucketAddresses["droplets"] = gcs.BackupBucketAddress{BucketName: firstBucketName, Path: "2006_01_02_15_04_05/droplets"}
				backupBucket.GetBlobReturns([]byte(`[{"name": "1970_01_01_00_00_00/droplets/file1"}]`), nil)
				backupBucket.DeleteBlobReturns(nil)
			})

			It("copies over all the common blobs from the previous backup", func() {
				err := backuper.CopyBlobsWithinBackupBucket(backupBucketAddresses)
				Expect(err).NotTo(HaveOccurred())

				Expect(backupBucket.GetBlobCallCount()).To(Equal(1))
				Expect(backupBucket.GetBlobArgsForCall(0)).To(Equal("2006_01_02_15_04_05/" + bucketPairID + "/common_blobs.json"))

				Expect(backupBucket.CopyBlobWithinBucketCallCount()).To(Equal(1))
				blob, path := backupBucket.CopyBlobWithinBucketArgsForCall(0)
				Expect(blob).To(Equal("1970_01_01_00_00_00/droplets/" + blob1))
				Expect(path).To(Equal("2006_01_02_15_04_05/" + bucketPairID + "/file1"))

				Expect(backupBucket.DeleteBlobCallCount()).To(Equal(1))
				Expect(backupBucket.DeleteBlobArgsForCall(0)).To(Equal("2006_01_02_15_04_05/" + bucketPairID + "/common_blobs.json"))
			})
		})

		Context("when the common_blobs.json cannot be read", func() {
			backupBucketAddresses := make(map[string]gcs.BackupBucketAddress)

			BeforeEach(func() {
				backupBucketAddresses["droplets"] = gcs.BackupBucketAddress{BucketName: firstBucketName, Path: "2006_01_02_15_04_05/droplets"}
				backupBucket.GetBlobReturns([]byte{}, fmt.Errorf("gcs error"))
			})

			It("returns a useful error", func() {
				err := backuper.CopyBlobsWithinBackupBucket(backupBucketAddresses)
				Expect(err).To(MatchError("failed to get 2006_01_02_15_04_05/droplets/common_blobs.json: gcs error"))
			})
		})

		Context("when the common_blobs.json is not valid JSON", func() {
			backupBucketAddresses := make(map[string]gcs.BackupBucketAddress)

			BeforeEach(func() {
				backupBucketAddresses["droplets"] = gcs.BackupBucketAddress{BucketName: firstBucketName, Path: "2006_01_02_15_04_05/droplets"}
				backupBucket.GetBlobReturns([]byte(`not json`), nil)
			})

			It("returns an error", func() {
				err := backuper.CopyBlobsWithinBackupBucket(backupBucketAddresses)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when deleting common_blobs.json faisl", func() {
			backupBucketAddresses := make(map[string]gcs.BackupBucketAddress)
			BeforeEach(func() {
				backupBucketAddresses["droplets"] = gcs.BackupBucketAddress{BucketName: firstBucketName, Path: "2006_01_02_15_04_05/droplets"}
				backupBucket.GetBlobReturns([]byte(`[{"name": "1970_01_01_00_00_00/droplets/file1"}]`), nil)
				backupBucket.DeleteBlobReturns(fmt.Errorf("I failed at deleting common_blobs.json"))
			})

			It("returns an error", func() {
				err := backuper.CopyBlobsWithinBackupBucket(backupBucketAddresses)
				Expect(err).To(MatchError("I failed at deleting common_blobs.json"))
			})
		})

		Context("when a common blob is missing", func() {
			backupBucketAddresses := make(map[string]gcs.BackupBucketAddress)

			BeforeEach(func() {
				backupBucketAddresses["droplets"] = gcs.BackupBucketAddress{BucketName: firstBucketName, Path: "2006_01_02_15_04_05/droplets"}
				backupBucket.GetBlobReturns([]byte(`[{"name": "1970_01_01_00_00_00/droplets/file1"}]`), nil)
				backupBucket.CopyBlobWithinBucketReturns(fmt.Errorf("gcs copy error"))
			})

			It("returns an error", func() {
				err := backuper.CopyBlobsWithinBackupBucket(backupBucketAddresses)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
