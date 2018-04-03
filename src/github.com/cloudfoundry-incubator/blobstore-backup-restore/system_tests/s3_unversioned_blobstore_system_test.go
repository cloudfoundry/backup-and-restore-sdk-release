package system_tests

import (
	"time"

	"strconv"

	"io/ioutil"

	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("S3 unversioned backuper", func() {
	var region string
	var bucket string
	var backupRegion string
	var backupBucket string
	var instanceArtifactDirPath string

	var blobKey string
	var localArtifact *os.File
	var unversionedBackuperInstance JobInstance

	BeforeEach(func() {
		unversionedBackuperInstance = JobInstance{
			deployment:    MustHaveEnv("BOSH_DEPLOYMENT"),
			instance:      "s3-unversioned-backuper",
			instanceIndex: "0",
		}

		region = MustHaveEnv("S3_UNVERSIONED_BUCKET_REGION")
		bucket = MustHaveEnv("S3_UNVERSIONED_BUCKET_NAME")

		backupRegion = MustHaveEnv("S3_UNVERSIONED_BACKUP_BUCKET_REGION")
		backupBucket = MustHaveEnv("S3_UNVERSIONED_BACKUP_BUCKET_NAME")

		instanceArtifactDirPath = "/tmp/s3-unversioned-blobstore-backup-restorer" + strconv.FormatInt(time.Now().Unix(), 10)
		unversionedBackuperInstance.runOnVMAndSucceed("mkdir -p " + instanceArtifactDirPath)
		var err error
		localArtifact, err = ioutil.TempFile("", "blobstore-")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		unversionedBackuperInstance.runOnVMAndSucceed("rm -rf " + instanceArtifactDirPath)
		err := os.Remove(localArtifact.Name())
		Expect(err).NotTo(HaveOccurred())
		deleteAllFilesFromBucket(region, bucket)
		deleteAllFilesFromBucket(backupRegion, backupBucket)
	})

	It("backs up from the source bucket to the backup bucket", func() {
		blobKey = uploadTimestampedFileToBucket(region, bucket, "some/folder/file1", "FILE1")

		unversionedBackuperInstance.runOnVMAndSucceed("BBR_ARTIFACT_DIRECTORY=" + instanceArtifactDirPath +
			" /var/vcap/jobs/s3-unversioned-blobstore-backup-restorer/bin/bbr/backup")

		filesList := listFilesFromBucket(backupRegion, backupBucket)

		Expect(filesList).To(ConsistOf(MatchRegexp(
			"\\d{4}_\\d{2}_\\d{2}_\\d{2}_\\d{2}_\\d{2}/my_bucket/" + blobKey + "$")))

		Expect(getFileContentsFromBucket(backupRegion, backupBucket, filesList[0])).To(Equal("FILE1"))

		session := unversionedBackuperInstance.downloadFromInstance(
			instanceArtifactDirPath+"/blobstore.json", localArtifact.Name())
		Expect(session).Should(gexec.Exit(0))
		fileContents, err := ioutil.ReadFile(localArtifact.Name())
		Expect(err).NotTo(HaveOccurred())
		Expect(fileContents).To(ContainSubstring("\"my_bucket\": {"))
		Expect(fileContents).To(ContainSubstring("\"bucket_name\": \"" + backupBucket + "\""))
		Expect(fileContents).To(ContainSubstring("\"bucket_region\": \"" + backupRegion + "\""))
		Expect(fileContents).To(MatchRegexp(
			"\"path\": \"\\d{4}_\\d{2}_\\d{2}_\\d{2}_\\d{2}_\\d{2}\\/my_bucket\""))
	})

	PIt("connects with a blobstore with custom CA cert", func() {

	})
})
