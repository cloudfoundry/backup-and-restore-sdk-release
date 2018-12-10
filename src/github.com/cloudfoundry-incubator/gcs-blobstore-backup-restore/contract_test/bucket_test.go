package contract_test

import (
	"fmt"

	"github.com/cloudfoundry-incubator/gcs-blobstore-backup-restore"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Bucket", func() {
	Describe("BuildBucketPairs", func() {
		It("builds bucket pairs", func() {
			config := map[string]gcs.Config{
				"droplets": {
					BucketName:       "droplets-bucket",
					BackupBucketName: "backup-droplets-bucket",
				},
			}

			buckets, err := gcs.BuildBucketPairs(MustHaveEnv("GCP_SERVICE_ACCOUNT_KEY"), config)
			Expect(err).NotTo(HaveOccurred())

			Expect(buckets).To(HaveLen(1))
			Expect(buckets["droplets"].LiveBucket.Name()).To(Equal("droplets-bucket"))
			Expect(buckets["droplets"].BackupBucket.Name()).To(Equal("backup-droplets-bucket"))
			Expect(buckets["droplets"].BackupFinder).NotTo(BeNil())
		})

		Context("when providing invalid service account key", func() {
			It("returns an error", func() {
				config := map[string]gcs.Config{
					"droplets": {
						BucketName:       "droplets-bucket",
						BackupBucketName: "backup-droplets-bucket",
					},
				}

				_, err := gcs.BuildBucketPairs("not-valid-json", config)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("ListBlobs", func() {
		var bucketName string
		var bucket gcs.Bucket
		var err error

		JustBeforeEach(func() {
			bucket, err = gcs.NewSDKBucket(MustHaveEnv("GCP_SERVICE_ACCOUNT_KEY"), bucketName)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when the bucket has files", func() {
			BeforeEach(func() {
				bucketName = CreateBucketWithTimestampedName("list_blobs")
				UploadFileWithDir(bucketName, "dir1", "file1", "file-content")
				UploadFileWithDir(bucketName, "dir2", "file2", "file-content")
				UploadFile(bucketName, "file3", "file-content")
			})

			It("lists all files", func() {
				blobs, err := bucket.ListBlobs()
				Expect(err).NotTo(HaveOccurred())
				Expect(blobs).To(ConsistOf(
					gcs.Blob{Name: "dir1/file1"},
					gcs.Blob{Name: "dir2/file2"},
					gcs.Blob{Name: "file3"},
				))
			})

			AfterEach(func() {
				DeleteBucket(bucketName)
			})
		})

		Context("when providing a non-existing bucket", func() {
			It("returns an error", func() {
				config := map[string]gcs.Config{
					"droplets": {
						BucketName:       "I-am-not-a-bucket",
						BackupBucketName: "definitely-not-a-bucket",
					},
				}

				bucketPair, err := gcs.BuildBucketPairs(MustHaveEnv("GCP_SERVICE_ACCOUNT_KEY"), config)
				Expect(err).NotTo(HaveOccurred())
				_, err = bucketPair["droplets"].LiveBucket.ListBlobs()
				Expect(err).To(MatchError("storage: bucket doesn't exist"))
			})
		})
	})

	Describe("LastBackupBlobs", func() {
		var backupBucketName string
		var backupBucket gcs.Bucket
		var err error

		JustBeforeEach(func() {
			backupBucket, err = gcs.NewSDKBucket(MustHaveEnv("GCP_SERVICE_ACCOUNT_KEY"), backupBucketName)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when the bucket exists", func() {
			BeforeEach(func() {
				backupBucketName = CreateBucketWithTimestampedName("list_last_backup_blobs")
			})

			AfterEach(func() {
				DeleteBucket(backupBucketName)
			})

			Context("when there are no previous backups", func() {
				It("returns an empty map", func() {
					blobs, err := backupBucket.LastBackupBlobs()
					Expect(err).NotTo(HaveOccurred())
					Expect(blobs).To(HaveLen(0))
				})
			})

			Context("when there is one previous backup", func() {

				BeforeEach(func() {
					UploadFileWithDir(backupBucketName, "1970_01_01_00_00_00", "file1", "file-content")
				})

				It("returns all the blobs from the previous backup", func() {
					blobs, err := backupBucket.LastBackupBlobs()
					Expect(err).NotTo(HaveOccurred())
					Expect(blobs).To(ConsistOf(gcs.Blob{Name: "1970_01_01_00_00_00/file1"}))
				})
			})

			Context("when there is more than one previous backup", func() {

				BeforeEach(func() {
					UploadFileWithDir(backupBucketName, "1970_01_01_00_00_00", "file1", "file-content1")
					UploadFileWithDir(backupBucketName, "1970_01_02_00_00_00", "file2", "file-content2")
				})

				It("returns only the blobs from the most recent previous backup", func() {
					blobs, err := backupBucket.LastBackupBlobs()
					Expect(err).NotTo(HaveOccurred())
					Expect(blobs).To(ConsistOf(gcs.Blob{Name: "1970_01_02_00_00_00/file2"}))
				})
			})
		})

		Context("when the bucket does not exist", func() {
			BeforeEach(func() {
				backupBucketName = "not-a-bucket"
			})

			It("returns an error", func() {
				_, err = backupBucket.LastBackupBlobs()
				Expect(err).To(MatchError("storage: bucket doesn't exist"))
			})
		})
	})

	Describe("CopyBlobWithinBucket", func() {
		var bucketName string
		var bucket gcs.Bucket
		var err error

		AfterEach(func() {
			DeleteBucket(bucketName)
		})

		Context("copying an existing file", func() {

			BeforeEach(func() {
				bucketName = CreateBucketWithTimestampedName("list_blobs")
				UploadFile(bucketName, "file1", "file-content")

				bucket, err = gcs.NewSDKBucket(MustHaveEnv("GCP_SERVICE_ACCOUNT_KEY"), bucketName)
				Expect(err).NotTo(HaveOccurred())
			})

			It("copies the blob to the specified location", func() {
				blob := gcs.Blob{Name: "file1"}

				err := bucket.CopyBlobWithinBucket(blob.Name, "copydir/file1")
				Expect(err).NotTo(HaveOccurred())

				blobs, err := bucket.ListBlobs()
				Expect(err).NotTo(HaveOccurred())
				Expect(blobs).To(ConsistOf(
					blob,
					gcs.Blob{Name: "copydir/file1"},
				))
			})
		})
	})

	Describe("CopyBlobBetweenBuckets", func() {
		var srcBucketName string
		var dstBucketName string
		var srcBucket gcs.Bucket
		var dstBucket gcs.Bucket
		var err error

		Context("copying an existing file", func() {

			BeforeEach(func() {
				srcBucketName = CreateBucketWithTimestampedName("src")
				dstBucketName = CreateBucketWithTimestampedName("dst")
				UploadFile(srcBucketName, "file1", "file-content")

				srcBucket, err = gcs.NewSDKBucket(MustHaveEnv("GCP_SERVICE_ACCOUNT_KEY"), srcBucketName)
				Expect(err).NotTo(HaveOccurred())

				dstBucket, err = gcs.NewSDKBucket(MustHaveEnv("GCP_SERVICE_ACCOUNT_KEY"), dstBucketName)
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				DeleteBucket(srcBucketName)
				DeleteBucket(dstBucketName)
			})

			It("copies the blob to the specified location", func() {
				blob := gcs.Blob{Name: "file1"}

				err := srcBucket.CopyBlobBetweenBuckets(dstBucket, blob.Name, "copydir/file1")
				Expect(err).NotTo(HaveOccurred())

				blobs, err := dstBucket.ListBlobs()
				Expect(err).NotTo(HaveOccurred())
				Expect(blobs).To(ConsistOf(
					gcs.Blob{Name: "copydir/file1"},
				))
			})
		})

		Context("copying a file that doesn't exist", func() {
			BeforeEach(func() {
				srcBucketName = CreateBucketWithTimestampedName("src")
				dstBucketName = CreateBucketWithTimestampedName("dst")

				srcBucket, err = gcs.NewSDKBucket(MustHaveEnv("GCP_SERVICE_ACCOUNT_KEY"), srcBucketName)
				Expect(err).NotTo(HaveOccurred())

				dstBucket, err = gcs.NewSDKBucket(MustHaveEnv("GCP_SERVICE_ACCOUNT_KEY"), dstBucketName)
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				DeleteBucket(srcBucketName)
				DeleteBucket(dstBucketName)
			})

			It("errors with a useful message", func() {
				err := srcBucket.CopyBlobBetweenBuckets(dstBucket, "foobar", "copydir/file1")
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError(ContainSubstring("failed to copy object: ")))
			})
		})

		Context("copying to a bucket that doesn't exist", func() {
			BeforeEach(func() {
				srcBucketName = CreateBucketWithTimestampedName("src")
				UploadFile(srcBucketName, "file1", "file-content")

				srcBucket, err = gcs.NewSDKBucket(MustHaveEnv("GCP_SERVICE_ACCOUNT_KEY"), srcBucketName)
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				DeleteBucket(srcBucketName)
			})

			It("errors", func() {
				err := srcBucket.CopyBlobBetweenBuckets(nil, "file1", "copydir/file1")
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("destination bucket does not exist"))
			})
		})
	})

	Describe("CopyBlobsBetweenBuckets", func() {
		var (
			srcBucketName string
			dstBucketName string
			srcBucket     gcs.Bucket
			dstBucket     gcs.Bucket
			badBucket     gcs.Bucket
			err           error
		)

		BeforeEach(func() {
			srcBucketName = CreateBucketWithTimestampedName("src")
			dstBucketName = CreateBucketWithTimestampedName("dst")
			UploadFile(srcBucketName, "file1", "file-content")
			UploadFile(dstBucketName, "file4", "file-content")
			UploadFileWithDir(srcBucketName, "sourcePath", "file2", "file-content2")
			UploadFileWithDir(srcBucketName, "sourcePath", "file3", "file-content3")

			srcBucket, err = gcs.NewSDKBucket(MustHaveEnv("GCP_SERVICE_ACCOUNT_KEY"), srcBucketName)
			Expect(err).NotTo(HaveOccurred())

			dstBucket, err = gcs.NewSDKBucket(MustHaveEnv("GCP_SERVICE_ACCOUNT_KEY"), dstBucketName)
			Expect(err).NotTo(HaveOccurred())

			badBucket, err = gcs.NewSDKBucket(MustHaveEnv("GCP_SERVICE_ACCOUNT_KEY"), "badBucket")
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			DeleteBucket(srcBucketName)
			DeleteBucket(dstBucketName)
		})

		It("copies only the blobs from source/sourcePath to destination/destinationPath", func() {
			err := srcBucket.CopyBlobsBetweenBuckets(dstBucket, "sourcePath")
			Expect(err).NotTo(HaveOccurred())

			blobs, err := dstBucket.ListBlobs()
			Expect(err).NotTo(HaveOccurred())
			Expect(blobs).To(ConsistOf(
				gcs.Blob{Name: "file2"},
				gcs.Blob{Name: "file3"},
				gcs.Blob{Name: "file4"},
			))
		})

		It("returns an error if the destination bucket does not exist", func() {
			err := srcBucket.CopyBlobsBetweenBuckets(nil, "sourcePath")
			Expect(err).To(MatchError("destination bucket does not exist"))
		})

		It("returns an error when the source bucket does not exist", func() {
			err := badBucket.CopyBlobsBetweenBuckets(dstBucket, "path")
			Expect(err).To(HaveOccurred())
		})

		It("returns an error when the destination bucket does not exist", func() {
			err := srcBucket.CopyBlobsBetweenBuckets(badBucket, "sourcePath")
			Expect(err).To(HaveOccurred())
		})

	})

	Describe("DeleteBlob", func() {
		var (
			bucketName string
			bucket     gcs.Bucket
			err        error
			dirName    = "mydir"
			fileName1  = "file1"
			fileName2  = "file2"
		)

		AfterEach(func() {
			DeleteBucket(bucketName)
		})

		Context("when deleting a file that doesn't exist", func() {
			BeforeEach(func() {
				bucketName = CreateBucketWithTimestampedName("src")
				UploadFile(bucketName, fileName1, "file-content")

				bucket, err = gcs.NewSDKBucket(MustHaveEnv("GCP_SERVICE_ACCOUNT_KEY"), bucketName)
				Expect(err).NotTo(HaveOccurred())
			})

			It("errors", func() {
				err := bucket.DeleteBlob(fileName1 + "idontexist")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when deleting existing files", func() {
			BeforeEach(func() {
				bucketName = CreateBucketWithTimestampedName("src")
				UploadFileWithDir(bucketName, dirName, fileName1, "file-content")
				UploadFileWithDir(bucketName, dirName, fileName2, "file-content")

				bucket, err = gcs.NewSDKBucket(MustHaveEnv("GCP_SERVICE_ACCOUNT_KEY"), bucketName)
				Expect(err).NotTo(HaveOccurred())
			})

			It("deletes all files and the folder", func() {
				err := bucket.DeleteBlob(fmt.Sprintf("%s/%s", dirName, fileName1))
				Expect(err).NotTo(HaveOccurred())

				err = bucket.DeleteBlob(fmt.Sprintf("%s/%s", dirName, fileName2))
				Expect(err).NotTo(HaveOccurred())

				blobs, err := bucket.ListBlobs()
				Expect(err).NotTo(HaveOccurred())
				Expect(blobs).To(BeNil())
			})
		})
	})

	Describe("CreateBackupCompleteBlob", func() {
		var (
			bucketName string
			bucket     gcs.Bucket
			err        error
		)

		BeforeEach(func() {
			bucketName = CreateBucketWithTimestampedName("backup-complete")

			bucket, err = gcs.NewSDKBucket(MustHaveEnv("GCP_SERVICE_ACCOUNT_KEY"), bucketName)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			DeleteBucket(bucketName)
		})

		Context("when creating the blob succeeds", func() {
			It("creates the blob correctly", func() {
				err := bucket.CreateBackupCompleteBlob("test-create-backup-complete")

				Expect(err).NotTo(HaveOccurred())
				blobs, err := bucket.ListBlobs()
				Expect(err).NotTo(HaveOccurred())
				Expect(blobs).To(ConsistOf(
					gcs.Blob{Name: "test-create-backup-complete/backup_complete"},
				))
			})
		})

		Context("when creating the blob fails", func() {
			It("returns the correct error", func() {
				bucket, err = gcs.NewSDKBucket(MustHaveEnv("GCP_SERVICE_ACCOUNT_KEY"), "iamnotabucket")
				Expect(err).NotTo(HaveOccurred())

				err := bucket.CreateBackupCompleteBlob("test-create-backup-complete")

				Expect(err).To(MatchError(ContainSubstring("failed creating backup complete blob")))
			})
		})
	})
})
