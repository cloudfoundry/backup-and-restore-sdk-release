package gcs_test

import (
	"strconv"
	"time"

	. "github.com/cloudfoundry-incubator/backup-and-restore-sdk-release-system-tests"
	. "github.com/onsi/ginkgo"
)

var _ = Describe("GCS Blobstore System Tests", func() {
	var (
		gcsClient               GCSClient
		bucket                  string
		blob1, blob2, blob3     string
		instance                JobInstance
		instanceArtifactDirPath string
	)

	BeforeEach(func() {
		bucket = MustHaveEnv("GCS_BUCKET_NAME")

		blob1 = timestampedName("test_file_1_")
		blob2 = timestampedName("test_file_2_")
		blob3 = timestampedName("test_file_3_")

		instance = JobInstance{
			Deployment: MustHaveEnv("BOSH_DEPLOYMENT"),
			Name:       "gcs-backuper",
			Index:      "0",
		}

		instanceArtifactDirPath = timestampedName("/tmp/gcs-blobstore-backup-restorer")
		instance.RunSuccessfully("mkdir -p " + instanceArtifactDirPath)
	})

	AfterEach(func() {
		gcsClient.DeleteBlobInBucket(bucket, blob1)
		gcsClient.DeleteBlobInBucket(bucket, blob2)
		gcsClient.DeleteBlobInBucket(bucket, blob3)
	})

	It("backs up a bucket", func() {
		gcsClient.WriteBlobToBucket(bucket, blob1, "TEST_BLOB_1")
		gcsClient.WriteBlobToBucket(bucket, blob2, "TEST_BLOB_2")
		gcsClient.WriteBlobToBucket(bucket, blob3, "TEST_BLOB_3")

		instance.RunSuccessfully("BBR_ARTIFACT_DIRECTORY=" + instanceArtifactDirPath + " /var/vcap/jobs/gcs-blobstore-backup-restorer/bin/bbr/backup")
	})
})

func timestampedName(prefix string) string {
	return prefix + strconv.FormatInt(time.Now().Unix(), 10)
}
