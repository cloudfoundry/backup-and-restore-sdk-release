package gcs_test

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/onsi/gomega/gexec"

	. "github.com/cloudfoundry-incubator/backup-and-restore-sdk-release-system-tests"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("GCS Blobstore System Tests", func() {
	var (
		gcsClient               GCSClient
		bucket, backupBucket    string
		blob1, blob2, blob3     string
		instance                JobInstance
		instanceArtifactDirPath string
	)

	BeforeEach(func() {
		bucket = MustHaveEnv("GCS_BUCKET_NAME")
		backupBucket = MustHaveEnv("GCS_BACKUP_BUCKET_NAME")

		blob1 = timestampedName("test_file_1_")
		blob2 = timestampedName("test_file_2_")
		blob3 = timestampedName("test_file_3_")

		instance = JobInstance{
			Deployment: MustHaveEnv("BOSH_DEPLOYMENT"),
			Name:       "gcs-backuper",
			Index:      "0",
		}

		instanceArtifactDirPath = "/var/vcap/store/gcs-blobstore-backup-restorer" + strconv.FormatInt(time.Now().Unix(), 10)
		instance.RunSuccessfully("sudo mkdir -p " + instanceArtifactDirPath)
	})

	AfterEach(func() {
		gcsClient.DeleteAllBlobInBucket(fmt.Sprintf(bucket + "/*"))
		gcsClient.DeleteAllBlobInBucket(fmt.Sprintf(backupBucket + "/*"))
	})

	Describe("Backup and bpm is enabled", func() {
		BeforeEach(func() {
			gcsClient.WriteBlobToBucket(bucket, blob1, "TEST_BLOB_1")
			gcsClient.WriteBlobToBucket(bucket, blob2, "TEST_BLOB_2")
			gcsClient.WriteBlobToBucket(bucket, blob3, "TEST_BLOB_3")
		})
		It("creates a backup and restores", func() {
			By("successfully running a backup", func() {
				instance.RunSuccessfully("sudo BBR_ARTIFACT_DIRECTORY=" + instanceArtifactDirPath +
					" /var/vcap/jobs/gcs-blobstore-backup-restorer/bin/bbr/backup")
			})

			By("creating a complete remote backup", func() {
				backupBucketFolders := gcsClient.ListDirsFromBucket(backupBucket)
				Expect(backupBucketFolders).To(MatchRegexp(
					".*\\d{4}_\\d{2}_\\d{2}_\\d{2}_\\d{2}_\\d{2}/"))

				backupBucketContent := getContentsOfBackupBucket(gcsClient, backupBucketFolders, "droplets")
				Expect(backupBucketContent).To(ContainSubstring(blob1))
				Expect(backupBucketContent).To(ContainSubstring(blob2))
				Expect(backupBucketContent).To(ContainSubstring(blob3))
			})

			By("generating a complete backup artifact", func() {
				session := instance.Run(fmt.Sprintf("cat %s/%s", instanceArtifactDirPath, "blobstore.json"))
				Expect(session).Should(gexec.Exit(0))
				fileContents := string(session.Out.Contents())

				Expect(fileContents).To(ContainSubstring("\"droplets\":{"))
				Expect(fileContents).To(ContainSubstring("\"bucket_name\":\"" + backupBucket + "\""))
				Expect(fileContents).To(MatchRegexp(
					"\"path\":\"\\d{4}_\\d{2}_\\d{2}_\\d{2}_\\d{2}_\\d{2}\\/droplets\""))
			})

			By("restoring from a backup artifact", func() {
				gcsClient.DeleteBlobInBucket(bucket, blob1)

				instance.RunSuccessfully("sudo BBR_ARTIFACT_DIRECTORY=" + instanceArtifactDirPath +
					" /var/vcap/jobs/gcs-blobstore-backup-restorer/bin/bbr/restore")

				liveBucketContent := gcsClient.ListDirsFromBucket(bucket)
				Expect(liveBucketContent).To(ContainSubstring(blob1))
				Expect(liveBucketContent).To(ContainSubstring(blob2))
				Expect(liveBucketContent).To(ContainSubstring(blob3))
			})
		})
	})
})

func timestampedName(prefix string) string {
	return prefix + strconv.FormatInt(time.Now().Unix(), 10)
}

func getContentsOfBackupBucket(gcsClient GCSClient, backupBucketTimestampedFolder, bucketID string) string {
	backupFolder := strings.TrimPrefix(backupBucketTimestampedFolder, "gs://")
	backupFolder = strings.TrimSuffix(backupFolder, "\n")
	backupFolder = backupFolder + bucketID
	return gcsClient.ListDirsFromBucket(backupFolder)
}
