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

			BeforeEach(func() {
				bucketName = CreateBucketWithTimestampedName("list_blobs")
				UploadFile(bucketName, "file1", "file-content")
				UploadFile(bucketName, "file2", "file-content")
				UploadFile(bucketName, "file3", "file-content")
			})

			It("lists all files", func() {
				blobs, err := bucket.ListBlobs()
				Expect(err).NotTo(HaveOccurred())
				Expect(blobs).To(ConsistOf(
					gcs.Blob{Name: "file1"},
					gcs.Blob{Name: "file2"},
					gcs.Blob{Name: "file3"},
				))
			})
		})

		Context("when the bucket has files in sub directories", func() {

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
				It("returns an empty map", func() {
					blobs, err := backupBucket.ListLastBackupBlobs()
					Expect(err).NotTo(HaveOccurred())
					Expect(blobs).To(HaveLen(0))
				})
			})

			Context("when there is one previous backup", func() {

				BeforeEach(func() {
					UploadFileWithDir(backupBucketName, "1970_01_01_00_00_00", "file1", "file-content")
				})

				It("returns all the blobs from the previous backup", func() {
					blobs, err := backupBucket.ListLastBackupBlobs()
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
					blobs, err := backupBucket.ListLastBackupBlobs()
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

		Context("copying a file that doesn't exist", func() {

			BeforeEach(func() {
				bucketName = CreateBucketWithTimestampedName("list_blobs")

				bucket, err = gcs.NewSDKBucket(MustHaveEnv("GCP_SERVICE_ACCOUNT_KEY"), bucketName)
				Expect(err).NotTo(HaveOccurred())
			})

			It("errors", func() {
				err := bucket.CopyBlobWithinBucket("foobar", "copydir/file1")
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
			bucketName  string
			bucket      gcs.Bucket
			err         error
			fileName    string
			fileContent []byte
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
			err = bucket.CreateFile(fileName, fileContent)
			Expect(err).NotTo(HaveOccurred())

			blobs, err := bucket.ListBlobs()
			Expect(err).NotTo(HaveOccurred())
			Expect(blobs).To(ConsistOf(
				gcs.Blob{Name: fileName},
			))
			Expect(ReadFile(bucketName, fileName)).To(Equal(string(fileContent)))
		})
	})

	Describe("GetBlob", func() {
		var (
			bucketName string
			bucket     gcs.Bucket
			err        error
			blobName   string
			content    string
		)

		BeforeEach(func() {
			bucketName = CreateBucketWithTimestampedName("get_blob")
			bucket, err = gcs.NewSDKBucket(MustHaveEnv("GCP_SERVICE_ACCOUNT_KEY"), bucketName)
			Expect(err).NotTo(HaveOccurred())

			blobName = "myBlob"
			content = "blobContent"
		})

		AfterEach(func() {
			DeleteBucket(bucketName)
		})

		Context("when the blob exists", func() {
			BeforeEach(func() {
				UploadFile(bucketName, blobName, content)
			})

			It("gets the blob content", func() {
				c, err := bucket.GetBlob(blobName)
				Expect(err).NotTo(HaveOccurred())
				Expect(string(c)).To(Equal(content))
			})
		})

		Context("when the blob does not exist", func() {
			It("returns an error", func() {
				_, err := bucket.GetBlob(blobName)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
