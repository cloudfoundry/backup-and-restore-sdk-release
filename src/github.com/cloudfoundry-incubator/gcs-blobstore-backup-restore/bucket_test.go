package gcs_test

import (
	"fmt"

	"github.com/cloudfoundry-incubator/gcs-blobstore-backup-restore"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Bucket", func() {
	Describe("BuildBuckets", func() {
		It("builds buckets", func() {
			config := map[string]gcs.Config{
				"droplets": {
					BucketName:       "droplets-bucket",
					BackupBucketName: "backup-droplets-bucket",
				},
			}

			buckets, err := gcs.BuildBuckets(MustHaveEnv("GCP_SERVICE_ACCOUNT_KEY"), config)
			Expect(err).NotTo(HaveOccurred())

			Expect(buckets).To(HaveLen(1))
			Expect(buckets["droplets"].Bucket.Name()).To(Equal("droplets-bucket"))
			Expect(buckets["droplets"].BackupBucket.Name()).To(Equal("backup-droplets-bucket"))
		})

		Context("when providing invalid service account key", func() {
			It("returns an error", func() {
				config := map[string]gcs.Config{
					"droplets": {
						BucketName:       "droplets-bucket",
						BackupBucketName: "backup-droplets-bucket",
					},
				}

				_, err := gcs.BuildBuckets("not-valid-json", config)
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

		AfterEach(func() {
			DeleteBucket(bucketName)
		})

		Context("when the bucket has a few files", func() {
			var file1GenerationID, file2GenerationID, file3GenerationID int64

			BeforeEach(func() {
				bucketName = CreateBucketWithTimestampedName("list_blobs")
				file1GenerationID = UploadFile(bucketName, "file1", "file-content")
				file2GenerationID = UploadFile(bucketName, "file2", "file-content")
				file3GenerationID = UploadFile(bucketName, "file3", "file-content")
			})

			It("lists all files and its generation_ids", func() {
				blobs, err := bucket.ListBlobs()
				Expect(err).NotTo(HaveOccurred())
				Expect(blobs).To(ConsistOf(
					gcs.Blob{Name: "file1", GenerationID: file1GenerationID},
					gcs.Blob{Name: "file2", GenerationID: file2GenerationID},
					gcs.Blob{Name: "file3", GenerationID: file3GenerationID},
				))
			})
		})

		Context("when the bucket has files in sub directories", func() {
			var file1GenerationID, file2GenerationID, file3GenerationID int64

			BeforeEach(func() {
				bucketName = CreateBucketWithTimestampedName("list_blobs")
				file1GenerationID = UploadFileWithDir(bucketName, "dir1", "file1", "file-content")
				file2GenerationID = UploadFileWithDir(bucketName, "dir2", "file2", "file-content")
				file3GenerationID = UploadFile(bucketName, "file3", "file-content")
			})

			It("lists all files and its generation_ids", func() {
				blobs, err := bucket.ListBlobs()
				Expect(err).NotTo(HaveOccurred())
				Expect(blobs).To(ConsistOf(
					gcs.Blob{Name: "dir1/file1", GenerationID: file1GenerationID},
					gcs.Blob{Name: "dir2/file2", GenerationID: file2GenerationID},
					gcs.Blob{Name: "file3", GenerationID: file3GenerationID},
				))
			})
		})
	})

	Describe("ListLastBackupBlobs", func() {
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
				It("returns an empty slice", func() {
					blobs, err := backupBucket.ListLastBackupBlobs()
					Expect(err).NotTo(HaveOccurred())
					Expect(blobs).To(HaveLen(0))
				})
			})

			Context("when there is one previous backup", func() {
				var file1GenerationID int64

				BeforeEach(func() {
					file1GenerationID = UploadFileWithDir(backupBucketName, "1970_01_01_00_00_00", "file1", "file-content")
				})

				It("returns all the blobs from the previous backup", func() {
					blobs, err := backupBucket.ListLastBackupBlobs()
					Expect(err).NotTo(HaveOccurred())
					Expect(blobs).To(ConsistOf(gcs.Blob{Name: "1970_01_01_00_00_00/file1", GenerationID: file1GenerationID}))
				})
			})

			Context("when there is more than one previous backup", func() {
				var file1GenerationID, file2GenerationID int64

				BeforeEach(func() {
					file1GenerationID = UploadFileWithDir(backupBucketName, "1970_01_01_00_00_00", "file1", "file-content1")
					file2GenerationID = UploadFileWithDir(backupBucketName, "1970_01_02_00_00_00", "file2", "file-content2")
				})

				It("returns only the blobs from the most recent previous backup", func() {
					blobs, err := backupBucket.ListLastBackupBlobs()
					Expect(err).NotTo(HaveOccurred())
					Expect(blobs).To(ConsistOf(gcs.Blob{Name: "1970_01_02_00_00_00/file2", GenerationID: file2GenerationID}))
				})
			})
		})

		Context("when the bucket does not exist", func() {
			BeforeEach(func() {
				backupBucketName = "not-a-bucket"
			})

			It("returns an error", func() {
				_, err = backupBucket.ListLastBackupBlobs()
				Expect(err).To(HaveOccurred())
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
			var file1GenerationID int64

			BeforeEach(func() {
				bucketName = CreateBucketWithTimestampedName("list_blobs")
				file1GenerationID = UploadFile(bucketName, "file1", "file-content")

				bucket, err = gcs.NewSDKBucket(MustHaveEnv("GCP_SERVICE_ACCOUNT_KEY"), bucketName)
				Expect(err).NotTo(HaveOccurred())
			})

			It("copies the blob to the specified location", func() {
				blob := gcs.Blob{Name: "file1", GenerationID: file1GenerationID}

				generatedID, err := bucket.CopyBlobWithinBucket(blob.Name, "copydir/file1")
				Expect(err).NotTo(HaveOccurred())

				blobs, err := bucket.ListBlobs()
				Expect(err).NotTo(HaveOccurred())
				Expect(blobs).To(ConsistOf(
					blob,
					gcs.Blob{Name: "copydir/file1", GenerationID: generatedID},
				))
			})
		})

		Context("copying a file that doesn't exist", func() {

			BeforeEach(func() {
				bucketName = CreateBucketWithTimestampedName("list_blobs")

				bucket, err = gcs.NewSDKBucket(MustHaveEnv("GCP_SERVICE_ACCOUNT_KEY"), bucketName)
				Expect(err).NotTo(HaveOccurred())
			})

			It("errors", func() {

				_, err := bucket.CopyBlobWithinBucket("foobar", "copydir/file1")
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("CopyBlobBetweenBuckets", func() {
		var srcBucketName string
		var dstBucketName string
		var srcBucket gcs.Bucket
		var dstBucket gcs.Bucket
		var err error
		var file1GenerationID int64

		Context("copying an existing file", func() {

			BeforeEach(func() {
				srcBucketName = CreateBucketWithTimestampedName("src")
				dstBucketName = CreateBucketWithTimestampedName("dst")
				file1GenerationID = UploadFile(srcBucketName, "file1", "file-content")

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
				blob := gcs.Blob{Name: "file1", GenerationID: file1GenerationID}

				generatedID, err := srcBucket.CopyBlobBetweenBuckets(dstBucket, blob.Name, "copydir/file1")
				Expect(err).NotTo(HaveOccurred())

				blobs, err := dstBucket.ListBlobs()
				Expect(err).NotTo(HaveOccurred())
				Expect(blobs).To(ConsistOf(
					gcs.Blob{Name: "copydir/file1", GenerationID: generatedID},
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

			It("errors", func() {
				_, err := srcBucket.CopyBlobBetweenBuckets(dstBucket, "foobar", "copydir/file1")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("copying to a bucket that doesn't exist", func() {
			BeforeEach(func() {
				srcBucketName = CreateBucketWithTimestampedName("src")
				file1GenerationID = UploadFile(srcBucketName, "file1", "file-content")

				srcBucket, err = gcs.NewSDKBucket(MustHaveEnv("GCP_SERVICE_ACCOUNT_KEY"), srcBucketName)
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				DeleteBucket(srcBucketName)
			})

			It("errors", func() {
				_, err := srcBucket.CopyBlobBetweenBuckets(nil, "file1", "copydir/file1")
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("destination bucket does not exist"))
			})
		})
	})

	Describe("Delete", func() {
		var bucketName string
		var bucket gcs.Bucket
		var err error
		var fileName = "file1"

		Context("deleting an existing file", func() {

			BeforeEach(func() {
				bucketName = CreateBucketWithTimestampedName("src")
				UploadFile(bucketName, fileName, "file-content")

				bucket, err = gcs.NewSDKBucket(MustHaveEnv("GCP_SERVICE_ACCOUNT_KEY"), bucketName)
				Expect(err).NotTo(HaveOccurred())

			})

			AfterEach(func() {
				DeleteBucket(bucketName)
			})

			It("deletes the blob", func() {

				err := bucket.DeleteBlob(fileName)
				Expect(err).NotTo(HaveOccurred())

				blobs, err := bucket.ListBlobs()
				Expect(err).NotTo(HaveOccurred())
				Expect(blobs).To(BeNil())
			})
		})

		Context("deleting a file that doesn't exist", func() {

			BeforeEach(func() {
				bucketName = CreateBucketWithTimestampedName("src")
				UploadFile(bucketName, fileName, "file-content")

				bucket, err = gcs.NewSDKBucket(MustHaveEnv("GCP_SERVICE_ACCOUNT_KEY"), bucketName)
				Expect(err).NotTo(HaveOccurred())

			})

			AfterEach(func() {
				DeleteBucket(bucketName)
			})

			It("errors", func() {

				err := bucket.DeleteBlob(fileName + "idontexist")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("deleting files in a folder", func() {
			var dirName = "mydir"
			var fileName1 = "file1"
			var fileName2 = "file2"
			BeforeEach(func() {
				bucketName = CreateBucketWithTimestampedName("src")
				UploadFileWithDir(bucketName, dirName, fileName1, "file-content")
				UploadFileWithDir(bucketName, dirName, fileName2, "file-content")

				bucket, err = gcs.NewSDKBucket(MustHaveEnv("GCP_SERVICE_ACCOUNT_KEY"), bucketName)
				Expect(err).NotTo(HaveOccurred())

			})

			AfterEach(func() {
				DeleteBucket(bucketName)
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

	Describe("CreateFile", func() {
		var (
			bucketName   string
			bucket       gcs.Bucket
			err          error
			generationID int64
			fileName     string
			fileContent  []byte
		)

		BeforeEach(func() {
			bucketName = CreateBucketWithTimestampedName("create_file")
			bucket, err = gcs.NewSDKBucket(MustHaveEnv("GCP_SERVICE_ACCOUNT_KEY"), bucketName)
			Expect(err).NotTo(HaveOccurred())

			fileName = "file1"
			fileContent = []byte("file_content")
		})

		AfterEach(func() {
			DeleteBucket(bucketName)
		})

		It("creates the file", func() {
			generationID, err = bucket.CreateFile(fileName, fileContent)
			Expect(err).NotTo(HaveOccurred())

			blobs, err := bucket.ListBlobs()
			Expect(err).NotTo(HaveOccurred())
			Expect(blobs).To(ConsistOf(
				gcs.Blob{Name: fileName, GenerationID: generationID},
			))
			Expect(ReadFile(bucketName, fileName)).To(Equal(string(fileContent)))
		})
	})
})
