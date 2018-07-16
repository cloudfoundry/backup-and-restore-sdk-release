package gcs_test

import (
	"fmt"
	"time"

	"github.com/cloudfoundry-incubator/gcs-blobstore-backup-restore"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Bucket", func() {
	Describe("BuildBuckets", func() {
		It("builds buckets", func() {
			config := map[string]gcs.Config{
				"droplets": {
					Name:              "droplets-bucket",
					ServiceAccountKey: MustHaveEnv("GCP_SERVICE_ACCOUNT_KEY"),
				},
			}

			buckets, err := gcs.BuildBuckets(config)
			Expect(err).NotTo(HaveOccurred())

			Expect(buckets).To(HaveLen(1))
			Expect(buckets["droplets"].Name()).To(Equal("droplets-bucket"))
		})
	})

	Describe("VersioningEnabled", func() {
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

		Context("when versioning is enabled on the given bucket", func() {
			BeforeEach(func() {
				bucketName = CreateBucketWithTimestampedName("versioning_on", true)
			})

			It("returns true", func() {
				versioningEnabled, err := bucket.VersioningEnabled()

				Expect(err).NotTo(HaveOccurred())
				Expect(versioningEnabled).To(BeTrue())
			})
		})

		Context("when versioning is not enabled on the given bucket", func() {
			BeforeEach(func() {
				bucketName = CreateBucketWithTimestampedName("versioning_off", false)
			})

			It("returns false", func() {
				versioningEnabled, err := bucket.VersioningEnabled()

				Expect(err).NotTo(HaveOccurred())
				Expect(versioningEnabled).To(BeFalse())
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

	Describe("CopyVersion", func() {
		var (
			bucketName       string
			bucket           gcs.Bucket
			err              error
			blobName         string
			blobGenerationID int64
		)

		BeforeEach(func() {
			bucketName = CreateBucketWithTimestampedName("copy_version", true)
			blobName = fmt.Sprintf("blob_%d", time.Now().UnixNano())
			blobGenerationID = UploadFile(bucketName, blobName, "file-content")
			bucket, err = gcs.NewSDKBucket(MustHaveEnv("GCP_SERVICE_ACCOUNT_KEY"), bucketName)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			DeleteBucket(bucketName)
		})

		Context("when the version is the current version of the blob", func() {
			It("is a noop", func() {
				blob := gcs.Blob{
					Name:         blobName,
					GenerationID: blobGenerationID,
				}

				err = bucket.CopyVersion(blob)

				Expect(err).NotTo(HaveOccurred())
				versions := ListBlobVersions(bucketName, blobName)
				Expect(versions).To(ConsistOf(blobGenerationID))
			})
		})

		Context("when the version is not the current version of the blob", func() {
			BeforeEach(func() {
				UploadFile(bucketName, blobName, "new-file-content")
			})

			It("copies the blob version to be the latest", func() {
				blob := gcs.Blob{
					Name:         blobName,
					GenerationID: blobGenerationID,
				}

				err = bucket.CopyVersion(blob)

				Expect(err).NotTo(HaveOccurred())
				content := GetBlobContents(bucketName, blobName)
				Expect(content).To(Equal("file-content"))
			})
		})

		Context("when the blob version is not found", func() {
			It("returns an error", func() {
				blobGenerationID = blobGenerationID + 1
				blob := gcs.Blob{
					Name:         blobName,
					GenerationID: blobGenerationID,
				}

				err = bucket.CopyVersion(blob)

				Expect(err).To(MatchError(ContainSubstring(
					fmt.Sprintf("error copying blob 'gs://%s/%s#%d'", bucketName, blobName, blobGenerationID),
				)))
			})
		})

		Context("when the blob does not exist", func() {
			It("returns an error", func() {
				blobName = "not-a-blob"
				blob := gcs.Blob{
					Name:         blobName,
					GenerationID: blobGenerationID,
				}

				err = bucket.CopyVersion(blob)

				Expect(err).To(MatchError(ContainSubstring(
					fmt.Sprintf("error getting blob attributes 'gs://%s/%s#%d'", bucketName, blobName, blobGenerationID),
				)))
			})
		})
	})
})
