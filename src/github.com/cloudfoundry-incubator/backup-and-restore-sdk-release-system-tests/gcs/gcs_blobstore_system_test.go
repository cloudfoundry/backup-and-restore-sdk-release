package gcs_test

import (
	"strconv"
	"time"

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

		gcsClient.DeleteBlobInBucket(backupBucket, blob1)
		gcsClient.DeleteBlobInBucket(backupBucket, blob2)
		gcsClient.DeleteBlobInBucket(backupBucket, blob3)
	})

	Context("when restoring to the same bucket", func() {
		Context("and no previous backup has been taken", func() {
			It("backs up and restores a bucket", func() {
				By("Creating blobs")
				gcsClient.WriteBlobToBucket(bucket, blob1, "TEST_BLOB_1")
				gcsClient.WriteBlobToBucket(bucket, blob2, "TEST_BLOB_2")
				gcsClient.WriteBlobToBucket(bucket, blob3, "TEST_BLOB_3")

				By("Running a backup")
				instance.RunSuccessfully("BBR_ARTIFACT_DIRECTORY=" + instanceArtifactDirPath + " /var/vcap/jobs/gcs-blobstore-backup-restorer/bin/bbr/backup")

				By("Seeing the live blob snapshot inside the live bucket")
				liveBucketContent := gcsClient.ListDirsFromBucket(bucket)
				Expect(liveBucketContent).To(ContainSubstring("temporary-backup-artifact"))

				By("Running unlock")
				instance.RunSuccessfully("BBR_ARTIFACT_DIRECTORY=" + instanceArtifactDirPath + " /var/vcap/jobs/gcs-blobstore-backup-restorer/bin/bbr/post-backup-unlock")

				By("Having a complete backup")
				liveBucketContent = gcsClient.ListDirsFromBucket(bucket)
				backupBucketContent := gcsClient.ListDirsFromBucket(backupBucket)
				Expect(backupBucketContent).To(Equal(liveBucketContent))

				By("Having cleaned up the live bucket")
				Expect(liveBucketContent).NotTo(ContainSubstring("temporary-backup-artifact"))

				By("Modifying the live bucket")
				gcsClient.WriteBlobToBucket(bucket, blob1, "TEST_BLOB_1_NEW")
				gcsClient.DeleteBlobInBucket(bucket, blob2)

				By("Running a restore")
				instance.RunSuccessfully("BBR_ARTIFACT_DIRECTORY=" + instanceArtifactDirPath + " /var/vcap/jobs/gcs-blobstore-backup-restorer/bin/bbr/restore")

				By("Seeing the blobs restored")
				Expect(gcsClient.ReadBlobFromBucket(bucket, blob1)).To(Equal("TEST_BLOB_1"))
				Expect(gcsClient.ReadBlobFromBucket(bucket, blob2)).To(Equal("TEST_BLOB_2"))
				Expect(gcsClient.ReadBlobFromBucket(bucket, blob3)).To(Equal("TEST_BLOB_3"))
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
