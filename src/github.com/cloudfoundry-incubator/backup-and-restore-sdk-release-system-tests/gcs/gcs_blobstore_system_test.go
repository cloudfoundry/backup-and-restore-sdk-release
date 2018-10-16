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
		gcsClient.DeleteBlobInBucket(bucket, blob1)
		gcsClient.DeleteBlobInBucket(bucket, blob2)
		gcsClient.DeleteBlobInBucket(bucket, blob3)

		gcsClient.DeleteAllBlobInBucket(fmt.Sprintf(backupBucket + "/*"))

	})

	Describe("Backup", func() {
		Context("When no previous backup has been taken", func() {
			It("creates a backup", func() {
				By("Creating blobs", func() {
					gcsClient.WriteBlobToBucket(bucket, blob1, "TEST_BLOB_1")
					gcsClient.WriteBlobToBucket(bucket, blob2, "TEST_BLOB_2")
					gcsClient.WriteBlobToBucket(bucket, blob3, "TEST_BLOB_3")
				})

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

					backupFolder := strings.TrimPrefix(backupBucketFolders, "gs://")
					backupFolder = strings.TrimSuffix(backupFolder, "\n")
					backupFolder = backupFolder + "droplets"
					backupBucketContent := gcsClient.ListDirsFromBucket(backupFolder)
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
						"\"path\":\"%s/\\d{4}_\\d{2}_\\d{2}_\\d{2}_\\d{2}_\\d{2}\\/droplets\"", backupBucket))
				})

				By("Having cleaned up the live bucket", func() {
					liveBucketContent := gcsClient.ListDirsFromBucket(bucket)
					Expect(liveBucketContent).NotTo(ContainSubstring("temporary-backup-artifact"))
				})
			})
		})

		Context("and a previous backup has been taken", func() {

		})
	})
	//
	//Context("when restoring to a clone bucket", func() {
	//	var cloneInstance JobInstance
	//	var cloneBucket string
	//
	//	BeforeEach(func() {
	//		cloneInstance = JobInstance{
	//			Deployment: MustHaveEnv("BOSH_DEPLOYMENT"),
	//			Name:       "gcs-restore-to-clone-bucket",
	//			Index:      "0",
	//		}
	//		cloneInstance.RunSuccessfully("mkdir -p " + instanceArtifactDirPath)
	//
	//		cloneBucket = MustHaveEnv("GCS_CLONE_BUCKET_NAME")
	//	})
	//	It("backs up and restores successfully", func() {
	//		gcsClient.WriteBlobToBucket(bucket, blob1, "TEST_BLOB_1")
	//		gcsClient.WriteBlobToBucket(bucket, blob2, "TEST_BLOB_2")
	//		gcsClient.WriteBlobToBucket(bucket, blob3, "TEST_BLOB_3")
	//
	//		instance.RunSuccessfully("BBR_ARTIFACT_DIRECTORY=" + instanceArtifactDirPath + " /var/vcap/jobs/gcs-blobstore-backup-restorer/bin/bbr/backup")
	//
	//		instance.Download(instanceArtifactDirPath+"/blobstore.json", "/tmp/blobstore.json")
	//		cloneInstance.Upload("/tmp/blobstore.json", instanceArtifactDirPath+"/blobstore.json")
	//
	//		cloneInstance.RunSuccessfully("BBR_ARTIFACT_DIRECTORY=" + instanceArtifactDirPath + " /var/vcap/jobs/gcs-blobstore-backup-restorer/bin/bbr/restore")
	//
	//		Expect(gcsClient.ReadBlobFromBucket(cloneBucket, blob1)).To(Equal("TEST_BLOB_1"))
	//		Expect(gcsClient.ReadBlobFromBucket(cloneBucket, blob2)).To(Equal("TEST_BLOB_2"))
	//		Expect(gcsClient.ReadBlobFromBucket(cloneBucket, blob3)).To(Equal("TEST_BLOB_3"))
	//	})
	//
	//	AfterEach(func() {
	//		gcsClient.DeleteBlobInBucket(cloneBucket, blob1)
	//		gcsClient.DeleteBlobInBucket(cloneBucket, blob2)
	//		gcsClient.DeleteBlobInBucket(cloneBucket, blob3)
	//	})
	//})

})

func timestampedName(prefix string) string {
	return prefix + strconv.FormatInt(time.Now().Unix(), 10)
}
