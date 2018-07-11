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
					Name:              "droplets-bucket",
					ServiceAccountKey: serviceAccountKeyJson,
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
			bucket, err = gcs.NewSDKBucket(serviceAccountKeyJson, bucketName)
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
			bucket, err = gcs.NewSDKBucket(serviceAccountKeyJson, bucketName)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			DeleteBucket(bucketName)
		})

		Context("when the bucket have a few files", func() {
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
})
