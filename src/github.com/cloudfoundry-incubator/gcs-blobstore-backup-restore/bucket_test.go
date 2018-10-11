package gcs_test

import (
	"github.com/cloudfoundry-incubator/gcs-blobstore-backup-restore"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Bucket", func() {
	Describe("BuildBuckets", func() {
		It("builds buckets", func() {
			config := map[string]gcs.Config{
				"droplets": {
					BucketName:        "droplets-bucket",
					BackupBucketName:  "backup-droplets-bucket",
					ServiceAccountKey: MustHaveEnv("GCP_SERVICE_ACCOUNT_KEY"),
				},
			}

			buckets, err := gcs.BuildBuckets(config)
			Expect(err).NotTo(HaveOccurred())

			Expect(buckets).To(HaveLen(1))
			Expect(buckets["droplets"].Bucket.Name()).To(Equal("droplets-bucket"))
			Expect(buckets["droplets"].BackupBucket.Name()).To(Equal("backup-droplets-bucket"))
		})

		Context("when providing invalid service account key", func() {
			It("returns an error", func() {
				config := map[string]gcs.Config{
					"droplets": {
						BucketName:        "droplets-bucket",
						BackupBucketName:  "backup-droplets-bucket",
						ServiceAccountKey: "this is not valid json",
					},
				}

				_, err := gcs.BuildBuckets(config)
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
				bucketName = CreateBucketWithTimestampedName("list_blobs", true)
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
	})

	Describe("CopyBlob", func() {
		var bucketName string
		var bucket gcs.Bucket
		var err error

		AfterEach(func() {
			DeleteBucket(bucketName)
		})

		Context("copying an existing file", func() {
			var file1GenerationID int64

			BeforeEach(func() {
				bucketName = CreateBucketWithTimestampedName("list_blobs", true)
				file1GenerationID = UploadFile(bucketName, "file1", "file-content")

				bucket, err = gcs.NewSDKBucket(MustHaveEnv("GCP_SERVICE_ACCOUNT_KEY"), bucketName)
				Expect(err).NotTo(HaveOccurred())
			})

			It("copies the blob to the specified location", func() {
				blob := gcs.Blob{Name: "file1", GenerationID: file1GenerationID}

				generatedID, err := bucket.CopyBlob(blob.Name, "copydir/file1")
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
				bucketName = CreateBucketWithTimestampedName("list_blobs", true)

				bucket, err = gcs.NewSDKBucket(MustHaveEnv("GCP_SERVICE_ACCOUNT_KEY"), bucketName)
				Expect(err).NotTo(HaveOccurred())
			})

			It("errors", func() {

				_, err := bucket.CopyBlob("foobar", "copydir/file1")
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
