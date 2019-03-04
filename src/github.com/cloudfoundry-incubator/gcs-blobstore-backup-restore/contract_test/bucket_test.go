package contract_test

import (
	"fmt"

	"github.com/cloudfoundry-incubator/gcs-blobstore-backup-restore/fakes"

	"github.com/cloudfoundry-incubator/gcs-blobstore-backup-restore"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Bucket", func() {
	Describe("BuildBackupsToComplete", func() {
		It("builds bucket pairs", func() {
			config := map[string]gcs.Config{
				"droplets": {
					BucketName:       "droplets-bucket",
					BackupBucketName: "backup-droplets-bucket",
				},
			}

			backupsToComplete, err := gcs.BuildBackupsToComplete(MustHaveEnv("GCP_SERVICE_ACCOUNT_KEY"), config)
			Expect(err).NotTo(HaveOccurred())

			Expect(backupsToComplete).To(HaveLen(1))
			Expect(backupsToComplete["droplets"].BucketPair.LiveBucket.Name()).To(Equal("droplets-bucket"))
			Expect(backupsToComplete["droplets"].BucketPair.BackupBucket.Name()).To(Equal("backup-droplets-bucket"))
		})

		Context("when providing invalid service account key", func() {
			It("returns an error", func() {
				config := map[string]gcs.Config{
					"droplets": {
						BucketName:       "droplets-bucket",
						BackupBucketName: "backup-droplets-bucket",
					},
				}

				_, err := gcs.BuildBackupsToComplete("not-valid-json", config)
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when two bucket pairs point to the same bucket", func() {
			It("builds the correct bucket pairs", func() {
				config := map[string]gcs.Config{
					"bucket-1": {
						BucketName:       "common-bucket",
						BackupBucketName: "backup-common-bucket",
					},

					"bucket-2": {
						BucketName:       "common-bucket",
						BackupBucketName: "backup-common-bucket",
					},
				}

				backupsToComplete, err := gcs.BuildBackupsToComplete(MustHaveEnv("GCP_SERVICE_ACCOUNT_KEY"), config)
				Expect(err).NotTo(HaveOccurred())

				Expect(backupsToComplete).To(HaveLen(2))
				Expect(backupsToComplete["bucket-1"].BucketPair.LiveBucket.Name()).To(Equal("common-bucket"))
				Expect(backupsToComplete["bucket-1"].BucketPair.BackupBucket.Name()).To(Equal("backup-common-bucket"))
				Expect(backupsToComplete["bucket-1"].SameAsBucketID).To(BeEmpty())
				Expect(backupsToComplete["bucket-2"].SameAsBucketID).To(Equal("bucket-1"))
				Expect(backupsToComplete["bucket-2"].BucketPair).To(Equal(gcs.BucketPair{}))
			})
		})
	})

	Describe("MarkSameBackupsToComplete", func() {
		It("marks backups to complete that are the same as another bucket ID", func() {
			liveBucket1 := new(fakes.FakeBucket)
			liveBucket1.NameReturns("live-bucket-1")
			liveBucket3 := new(fakes.FakeBucket)
			liveBucket3.NameReturns("live-bucket-3")

			backupsToComplete := map[string]gcs.BackupToComplete{
				"bucket-4": {
					BucketPair: gcs.BucketPair{
						LiveBucket: liveBucket1,
					},
				},
				"bucket-2": {
					BucketPair: gcs.BucketPair{
						LiveBucket: liveBucket3,
					},
				},
				"bucket-1": {
					BucketPair: gcs.BucketPair{
						LiveBucket: liveBucket1,
					},
				},
				"bucket-3": {
					BucketPair: gcs.BucketPair{
						LiveBucket: liveBucket3,
					},
				},
			}

			markedSameBackupsToComplete := gcs.MarkSameBackupsToComplete(backupsToComplete)

			Expect(markedSameBackupsToComplete).To(Equal(
				map[string]gcs.BackupToComplete{
					"bucket-1": {
						BucketPair: gcs.BucketPair{
							LiveBucket: liveBucket1,
						},
					},
					"bucket-2": {
						BucketPair: gcs.BucketPair{
							LiveBucket: liveBucket3,
						},
					},
					"bucket-3": {
						SameAsBucketID: "bucket-2",
					},
					"bucket-4": {
						SameAsBucketID: "bucket-1",
					},
				},
			))
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

		Context("when the bucket contains multiple files and some match the prefix", func() {
			BeforeEach(func() {
				bucketName = CreateBucketWithTimestampedName("list_blobs")
				UploadFileWithDir(bucketName, "my/prefix", "file1", "file-content")
				UploadFileWithDir(bucketName, "not/my/prefix", "file2", "file-content")
				UploadFile(bucketName, "file3", "file-content")
			})

			It("lists all files that have the prefix", func() {
				blobs, err := bucket.ListBlobs("my/prefix")
				Expect(err).NotTo(HaveOccurred())
				Expect(blobs).To(ConsistOf(
					gcs.NewBlob("my/prefix/file1"),
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

				backupsToComplete, err := gcs.BuildBackupsToComplete(MustHaveEnv("GCP_SERVICE_ACCOUNT_KEY"), config)
				Expect(err).NotTo(HaveOccurred())
				_, err = backupsToComplete["droplets"].BucketPair.LiveBucket.ListBlobs("")
				Expect(err).To(MatchError("storage: bucket doesn't exist"))
			})
		})
	})

	Describe("CopyBlobToBucket", func() {
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
				blob := gcs.NewBlob("file1")

				err := srcBucket.CopyBlobToBucket(dstBucket, blob.Name(), "copydir/file1")
				Expect(err).NotTo(HaveOccurred())

				blobs, err := dstBucket.ListBlobs("")
				Expect(err).NotTo(HaveOccurred())
				Expect(blobs).To(ConsistOf(
					gcs.NewBlob("copydir/file1"),
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
				err := srcBucket.CopyBlobToBucket(dstBucket, "foobar", "copydir/file1")
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
				err := srcBucket.CopyBlobToBucket(nil, "file1", "copydir/file1")
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("destination bucket does not exist"))
			})
		})
	})

	Describe("CopyBlobsToBucket", func() {
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
			UploadFile(srcBucketName, "notInSourcePath", "file-content")
			UploadFile(dstBucketName, "alreadyInDstBucket", "file-content")
			UploadFileWithDir(srcBucketName, "sourcePath", "file1", "file-content1")
			UploadFileWithDir(srcBucketName, "sourcePath", "file2", "file-content2")

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
			err := srcBucket.CopyBlobsToBucket(dstBucket, "sourcePath")
			Expect(err).NotTo(HaveOccurred())

			blobs, err := dstBucket.ListBlobs("")
			Expect(err).NotTo(HaveOccurred())
			Expect(blobs).To(ConsistOf(
				gcs.NewBlob("file1"),
				gcs.NewBlob("file2"),
				gcs.NewBlob("alreadyInDstBucket"),
			))
		})

		It("returns an error if the destination bucket does not exist", func() {
			err := srcBucket.CopyBlobsToBucket(nil, "sourcePath")
			Expect(err).To(MatchError("destination bucket does not exist"))
		})

		It("returns an error when the source bucket does not exist", func() {
			err := badBucket.CopyBlobsToBucket(dstBucket, "path")
			Expect(err).To(HaveOccurred())
		})

		It("returns an error when the destination bucket does not exist", func() {
			err := srcBucket.CopyBlobsToBucket(badBucket, "sourcePath")
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

				blobs, err := bucket.ListBlobs("")
				Expect(err).NotTo(HaveOccurred())
				Expect(blobs).To(BeNil())
			})
		})
	})

})
