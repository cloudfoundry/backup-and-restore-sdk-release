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
		gcsClient                  GCSClient
		bucket, backupBucket       string
		blob1, blob2, blob3, blob4 string
		instance                   JobInstance
		instanceArtifactDirPath    string
	)

	BeforeEach(func() {
		bucket = MustHaveEnv("GCS_BUCKET_NAME")
		backupBucket = MustHaveEnv("GCS_BACKUP_BUCKET_NAME")

		blob1 = timestampedName("test_file_1_")
		blob2 = timestampedName("test_file_2_")
		blob3 = timestampedName("test_file_3_")
		blob4 = timestampedName("test_file_4_")

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
		Context("when no previous backup has been taken", func() {
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
					Expect(backupBucketContent).To(ContainSubstring("backup_complete"))
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

		Context("when a previous backup has been taken", func() {
			var previousBackupTimestamp = "1970_01_01_00_00_00"
			BeforeEach(func() {
				gcsClient.WriteBlobToBucket(bucket, blob1, "TEST_BLOB_1")
				gcsClient.WriteBlobToBucket(bucket, blob2, "TEST_BLOB_2")
				gcsClient.WriteBlobToBucket(bucket, blob3, "TEST_BLOB_3")
			})

			Context("and there are new blobs since the previous backup", func() {
				BeforeEach(func() {
					previousBackupDir := fmt.Sprintf("%s/droplets/", previousBackupTimestamp)
					gcsClient.WriteBlobToBucket(backupBucket, previousBackupDir+blob1, "TEST_BLOB_1")
					gcsClient.WriteBlobToBucket(backupBucket, previousBackupDir+blob2, "TEST_BLOB_2")
					gcsClient.WriteBlobToBucket(backupBucket, previousBackupDir+blob3, "TEST_BLOB_3")

					gcsClient.WriteBlobToBucket(bucket, blob4, "TEST_BLOB_4")
				})

				It("creates a complete backup and restores", func() {
					var backupBucketFolders string
					By("successfully running a backup", func() {
						instance.RunSuccessfully("sudo BBR_ARTIFACT_DIRECTORY=" +
							instanceArtifactDirPath + " /var/vcap/jobs/gcs-blobstore-backup-restorer/bin/bbr/backup")
					})

					By("not overwriting the previous backup artifact", func() {
						backupBucketFolders = gcsClient.ListDirsFromBucket(backupBucket)
						backupBucketFolders = strings.TrimSuffix(backupBucketFolders, "\n")
						backupDirs := strings.Split(backupBucketFolders, "\n")
						Expect(backupDirs).To(HaveLen(2))
					})

					By("creating a complete remote backup", func() {
						backupBucketFolders = removePreviousBackup(backupBucketFolders, backupBucket, previousBackupTimestamp)

						Expect(backupBucketFolders).To(MatchRegexp(
							".*\\d{4}_\\d{2}_\\d{2}_\\d{2}_\\d{2}_\\d{2}/"))

						backupBucketContent := getContentsOfBackupBucket(gcsClient, backupBucketFolders, "droplets")
						Expect(backupBucketContent).To(ContainSubstring(blob1))
						Expect(backupBucketContent).To(ContainSubstring(blob2))
						Expect(backupBucketContent).To(ContainSubstring(blob3))
						Expect(backupBucketContent).To(ContainSubstring(blob4))
					})

					By("restoring from the backup artifact", func() {
						gcsClient.DeleteBlobInBucket(bucket, blob1)

						instance.RunSuccessfully("sudo BBR_ARTIFACT_DIRECTORY=" + instanceArtifactDirPath +
							" /var/vcap/jobs/gcs-blobstore-backup-restorer/bin/bbr/restore")

						liveBucketContent := gcsClient.ListDirsFromBucket(bucket)
						Expect(liveBucketContent).To(ContainSubstring(blob1))
						Expect(liveBucketContent).To(ContainSubstring(blob2))
						Expect(liveBucketContent).To(ContainSubstring(blob3))
						Expect(liveBucketContent).To(ContainSubstring(blob4))
						Expect(liveBucketContent).NotTo(ContainSubstring("backup_complete"))
					})
				})
			})
		})
	})
})

func removePreviousBackup(backupBucketFolders, backupBucket, timestamp string) string {
	return strings.Replace(backupBucketFolders, fmt.Sprintf("gs://%s/%s/\n", backupBucket, timestamp), "", 1)
}

func timestampedName(prefix string) string {
	return prefix + strconv.FormatInt(time.Now().Unix(), 10)
}

func getContentsOfBackupBucket(gcsClient GCSClient, backupBucketTimestampedFolder, bucketID string) string {
	backupFolder := strings.TrimPrefix(backupBucketTimestampedFolder, "gs://")
	backupFolder = strings.TrimSuffix(backupFolder, "\n")
	backupFolder = backupFolder + bucketID
	return gcsClient.ListDirsFromBucket(backupFolder)
}
