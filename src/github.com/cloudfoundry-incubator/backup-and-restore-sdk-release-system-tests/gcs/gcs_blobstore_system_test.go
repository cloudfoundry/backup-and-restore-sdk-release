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

		instanceArtifactDirPath = timestampedName("/tmp/gcs-blobstore-backup-restorer")
		instance.RunSuccessfully("mkdir -p " + instanceArtifactDirPath)
	})

	AfterEach(func() {
		gcsClient.DeleteAllBlobInBucket(fmt.Sprintf(bucket + "/*"))
		gcsClient.DeleteAllBlobInBucket(fmt.Sprintf(backupBucket + "/*"))
	})

	Describe("Backup", func() {
		Context("When no previous backup has been taken", func() {
			BeforeEach(func() {
				gcsClient.WriteBlobToBucket(bucket, blob1, "TEST_BLOB_1")
				gcsClient.WriteBlobToBucket(bucket, blob2, "TEST_BLOB_2")
				gcsClient.WriteBlobToBucket(bucket, blob3, "TEST_BLOB_3")
			})
			It("creates a backup", func() {
				By("Running a backup", func() {
					instance.RunSuccessfully("BBR_ARTIFACT_DIRECTORY=" + instanceArtifactDirPath + " /var/vcap/jobs/gcs-blobstore-backup-restorer/bin/bbr/backup")
				})

				By("Seeing the live blob snapshot inside the live bucket", func() {
					liveBucketContent := gcsClient.ListDirsFromBucket(bucket)
					Expect(liveBucketContent).To(ContainSubstring("temporary-backup-artifact"))
					liveBucketContent = gcsClient.ListDirsFromBucket(bucket + "/temporary-backup-artifact")
					Expect(liveBucketContent).To(ContainSubstring("temporary-backup-artifact/" + blob1))
					Expect(liveBucketContent).To(ContainSubstring("temporary-backup-artifact/" + blob2))
					Expect(liveBucketContent).To(ContainSubstring("temporary-backup-artifact/" + blob3))
				})

				By("Running unlock", func() {
					instance.RunSuccessfully("BBR_ARTIFACT_DIRECTORY=" + instanceArtifactDirPath + " /var/vcap/jobs/gcs-blobstore-backup-restorer/bin/bbr/post-backup-unlock")
				})

				By("Having a complete remote backup", func() {
					backupBucketFolders := gcsClient.ListDirsFromBucket(backupBucket)
					Expect(backupBucketFolders).To(MatchRegexp(
						".*\\d{4}_\\d{2}_\\d{2}_\\d{2}_\\d{2}_\\d{2}/"))

					backupBucketContent := getContentsOfBackupBucket(gcsClient, backupBucketFolders, "droplets")
					Expect(backupBucketContent).To(ContainSubstring(blob1))
					Expect(backupBucketContent).To(ContainSubstring(blob2))
					Expect(backupBucketContent).To(ContainSubstring(blob3))
				})

				By("Having a complete backup artifact", func() {
					session := instance.Run(fmt.Sprintf("cat %s/%s", instanceArtifactDirPath, "blobstore.json"))
					Expect(session).Should(gexec.Exit(0))
					fileContents := string(session.Out.Contents())

					Expect(fileContents).To(ContainSubstring("\"droplets\":{"))
					Expect(fileContents).To(ContainSubstring("\"bucket_name\":\"" + backupBucket + "\""))
					Expect(fileContents).To(MatchRegexp(
						"\"path\":\"\\d{4}_\\d{2}_\\d{2}_\\d{2}_\\d{2}_\\d{2}\\/droplets\""))
				})

				By("Having cleaned up the live bucket", func() {
					liveBucketContent := gcsClient.ListDirsFromBucket(bucket)
					Expect(liveBucketContent).NotTo(ContainSubstring("temporary-backup-artifact"))
				})
			})
		})

		Context("and a previous backup has been taken", func() {
			var previousBackupTimestamp = "1970_01_01_00_00_00"
			BeforeEach(func() {
				gcsClient.WriteBlobToBucket(bucket, blob1, "TEST_BLOB_1")
				gcsClient.WriteBlobToBucket(bucket, blob2, "TEST_BLOB_2")
				gcsClient.WriteBlobToBucket(bucket, blob3, "TEST_BLOB_3")
			})

			Context("and there are no new blobs since the previous backup", func() {
				BeforeEach(func() {
					previousBackupDir := fmt.Sprintf("%s/droplets/", previousBackupTimestamp)
					gcsClient.WriteBlobToBucket(backupBucket, previousBackupDir+blob1, "TEST_BLOB_1")
					gcsClient.WriteBlobToBucket(backupBucket, previousBackupDir+blob2, "TEST_BLOB_2")
					gcsClient.WriteBlobToBucket(backupBucket, previousBackupDir+blob3, "TEST_BLOB_3")
				})

				It("creates a complete backup", func() {
					var backupBucketFolders string
					By("Running a backup", func() {
						instance.RunSuccessfully("BBR_ARTIFACT_DIRECTORY=" + instanceArtifactDirPath + " /var/vcap/jobs/gcs-blobstore-backup-restorer/bin/bbr/backup")
					})

					By("Seeing a snapshot inside the live bucket containing only a common_blobs.json", func() {
						liveBucketContent := gcsClient.ListDirsFromBucket(bucket)
						Expect(liveBucketContent).To(ContainSubstring("temporary-backup-artifact"))
						liveBucketContent = gcsClient.ListDirsFromBucket(bucket + "/temporary-backup-artifact")
						Expect(liveBucketContent).To(Equal("gs://" + bucket + "/temporary-backup-artifact/common_blobs.json\n"))
					})

					By("Running unlock", func() {
						instance.RunSuccessfully("BBR_ARTIFACT_DIRECTORY=" + instanceArtifactDirPath + " /var/vcap/jobs/gcs-blobstore-backup-restorer/bin/bbr/post-backup-unlock")
					})

					By("Not overwriting the previous backup artifact", func() {
						backupBucketFolders = gcsClient.ListDirsFromBucket(backupBucket)
						backupBucketFolders = strings.TrimSuffix(backupBucketFolders, "\n")
						backupDirs := strings.Split(backupBucketFolders, "\n")
						Expect(backupDirs).To(HaveLen(2))
					})

					By("Having a complete remote backup", func() {
						backupBucketFolders = removePreviousBackup(backupBucketFolders, backupBucket, previousBackupTimestamp)

						Expect(backupBucketFolders).To(MatchRegexp(
							".*\\d{4}_\\d{2}_\\d{2}_\\d{2}_\\d{2}_\\d{2}/"))

						backupBucketContent := getContentsOfBackupBucket(gcsClient, backupBucketFolders, "droplets")
						Expect(backupBucketContent).To(ContainSubstring(blob1))
						Expect(backupBucketContent).To(ContainSubstring(blob2))
						Expect(backupBucketContent).To(ContainSubstring(blob3))
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
