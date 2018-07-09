package gcp_test

import (
	"strconv"
	"time"

	. "github.com/cloudfoundry-incubator/backup-and-restore-sdk-release-system-tests"
	. "github.com/onsi/ginkgo"
)

var _ = Describe("GCPBlobstoreSystem", func() {
	var (
		gcpClient           GCPClient
		bucket              string
		blob1, blob2, blob3 string
	)

	BeforeEach(func() {
		bucket = MustHaveEnv("GCP_BUCKET_NAME")

		blob1 = blobName("test_file_1_")
		blob2 = blobName("test_file_2_")
		blob3 = blobName("test_file_3_")
	})

	AfterEach(func() {
		gcpClient.DeleteBlobInBucket(bucket, blob1)
		gcpClient.DeleteBlobInBucket(bucket, blob2)
		gcpClient.DeleteBlobInBucket(bucket, blob3)
	})

	It("backs up a bucket", func() {
		gcpClient.WriteBlobToBucket(bucket, blob1, "TEST_BLOB_1")
		gcpClient.WriteBlobToBucket(bucket, blob2, "TEST_BLOB_2")
		gcpClient.WriteBlobToBucket(bucket, blob3, "TEST_BLOB_3")
	})
})

func blobName(prefix string) string {
	return prefix + strconv.FormatInt(time.Now().Unix(), 10)
}
