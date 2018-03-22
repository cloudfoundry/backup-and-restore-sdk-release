package system_tests

import (
	"time"

	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = PDescribe("S3 unversioned backuper", func() {
	var region string
	var bucket string
	var backupRegion string
	var backupBucket string
	var artifactDirPath string

	var fileName1 string

	var unversionedBackuperInstance JobInstance

	BeforeEach(func() {
		unversionedBackuperInstance = JobInstance{
			deployment:    MustHaveEnv("BOSH_DEPLOYMENT"),
			instance:      "s3-unversioned-backuper",
			instanceIndex: "0",
		}

		region = MustHaveEnv("BUCKET_REGION")
		bucket = MustHaveEnv("BUCKET_NAME")

		backupRegion = MustHaveEnv("BACKUP_BUCKET_REGION")
		backupBucket = MustHaveEnv("BACKUP_BUCKET_NAME")

		artifactDirPath = "/tmp/s3-unversioned-blobstore-backup-restorer" + strconv.FormatInt(time.Now().Unix(), 10)
		unversionedBackuperInstance.runOnVMAndSucceed("mkdir -p " + artifactDirPath)
	})

	AfterEach(func() {
		unversionedBackuperInstance.runOnVMAndSucceed("rm -rf " + artifactDirPath)
	})

	It("backs up from the source bucket to the backup bucket", func() {
		fileName1 = uploadTimestampedFileToBucket(region, bucket, "file1", "FILE1")

		unversionedBackuperInstance.runOnVMAndSucceed("BBR_ARTIFACT_DIRECTORY=" + artifactDirPath +
			" /var/vcap/jobs/s3-unversioned-blobstore-backup-restorer/bin/bbr/backup")

		unversionedBackuperInstance.downloadFromInstance(artifactDirPath+"/blobstore.json", "/tmp/blobstore.json")

		filesList := listFilesFromBucket(backupRegion, backupBucket)
		Expect(filesList).To(ConsistOf(fileName1))

		Expect(getFileContentsFromBucket(backupRegion, backupBucket, fileName1)).To(Equal("FILE1"))
	})

	PIt("connects with a blobstore with custom CA cert", func() {

	})
})
