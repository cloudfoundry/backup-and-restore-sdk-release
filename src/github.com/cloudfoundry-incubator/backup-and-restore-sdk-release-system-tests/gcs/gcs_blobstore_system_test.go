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

var (
	gcsClient               GCSClient
	bucket, backupBucket    string
	instance                JobInstance
	instanceArtifactDirPath string
)

var _ = Describe("GCS Blobstore System Tests", func() {
	Describe("Backup and bpm is enabled", func() {
		BeforeEach(func() {
			bucket = MustHaveEnv("GCS_BUCKET_NAME")
			backupBucket = MustHaveEnv("GCS_BACKUP_BUCKET_NAME")
		})

		AfterEach(func() {
			gcsClient.DeleteAllBlobInBucket(fmt.Sprintf(bucket + "/*"))
			gcsClient.DeleteAllBlobInBucket(fmt.Sprintf(backupBucket + "/*"))
		})

		Context("when there is single live bucket", func() {
			BeforeEach(func() {
				instance = JobInstance{
					Deployment: MustHaveEnv("BOSH_DEPLOYMENT"),
					Name:       "gcs-backuper",
					Index:      "0",
				}

				instanceArtifactDirPath = "/var/vcap/store/gcs-blobstore-backup-restorer" + strconv.FormatInt(time.Now().Unix(), 10)
				instance.RunSuccessfully("sudo mkdir -p " + instanceArtifactDirPath)
			})
			Context("and there is large number files in the bucket", func() {
				numberOfBlobs := 2000
				BeforeEach(func() {
					gcsClient.WriteNBlobsToBucket(bucket, "test_file_%d_", "TEST_BLOB_%d", numberOfBlobs)
				})
				runTestWithBlobs(numberOfBlobs)
			})
			Context("and there is large file in the bucket", func() {
				sizeOfBlob := 10
				BeforeEach(func() {
					gcsClient.WriteNSizeBlobToBucket(bucket, "test_file_0_", sizeOfBlob)
				})
				runTestWithBlobs(1)
			})
		})
	})
})

func runTestWithBlobs(numberOfBlobs int) {
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
			for i := 0; i < numberOfBlobs; i++ {
				Expect(backupBucketContent).To(ContainSubstring(fmt.Sprintf("test_file_%d_", i)))
			}
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
			gcsClient.DeleteBlobInBucket(bucket, "**/test_file_0_*")

			instance.RunSuccessfully("sudo BBR_ARTIFACT_DIRECTORY=" + instanceArtifactDirPath +
				" /var/vcap/jobs/gcs-blobstore-backup-restorer/bin/bbr/restore")

			liveBucketContent := gcsClient.ListDirsFromBucket(bucket)
			for i := 0; i < numberOfBlobs; i++ {
				Expect(liveBucketContent).To(ContainSubstring(fmt.Sprintf("test_file_%d_", i)))
			}
		})
	})
}

func getContentsOfBackupBucket(gcsClient GCSClient, backupBucketTimestampedFolder, bucketID string) string {
	backupFolder := strings.TrimPrefix(backupBucketTimestampedFolder, "gs://")
	backupFolder = strings.TrimSuffix(backupFolder, "\n")
	backupFolder = backupFolder + bucketID
	return gcsClient.ListDirsFromBucket(backupFolder)
}
