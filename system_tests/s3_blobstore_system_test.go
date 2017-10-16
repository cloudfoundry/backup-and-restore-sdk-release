package system_tests

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"time"

	"strconv"

	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("S3 backuper", func() {
	var fileName1, fileVersion1 string
	var fileName2, fileVersion2 string
	var tmpDir string

	var backuperInstance JobInstance

	BeforeEach(func() {
		backuperInstance = JobInstance{
			deployment:    MustHaveEnv("BOSH_DEPLOYMENT"),
			instance:      "backuper",
			instanceIndex: "0",
		}

		fileName1, fileVersion1 = uploadEmptyTimestampedFile("file1")
		fileName2, fileVersion2 = uploadEmptyTimestampedFile("file2")

		tmpDir = "/tmp/aws-s3-versioned-blobstore-backup-restorer" + strconv.FormatInt(time.Now().Unix(), 10)
		backuperInstance.runOnVMAndSucceed("mkdir -p " + tmpDir)
	})

	AfterEach(func() {
		deleteVersion(fileName1, fileVersion1)
		deleteVersion(fileName2, fileVersion2)

		backuperInstance.runOnVMAndSucceed("rm -rf " + tmpDir)
	})

	It("saves the version metadata", func() {
		backuperInstance.runOnVMAndSucceed("BBR_ARTIFACT_DIRECTORY=" + tmpDir + " /var/vcap/jobs/aws-s3-versioned-blobstore-backup-restorer/bin/bbr/backup")

		session := backuperInstance.runOnInstance("cat " + tmpDir + "/blobstore.json")
		fileContents := session.Out.Contents()

		Expect(fileContents).To(ContainSubstring(fileName1))
		Expect(fileContents).To(ContainSubstring(fileName2))
		Expect(fileContents).To(ContainSubstring(fileVersion1))
		Expect(fileContents).To(ContainSubstring(fileVersion2))
		Expect(fileContents).To(ContainSubstring(MustHaveEnv("AWS_TEST_BUCKET_NAME")))
		Expect(fileContents).To(ContainSubstring(MustHaveEnv("AWS_TEST_BUCKET_REGION")))
	})
})

func uploadEmptyTimestampedFile(prefix string) (string, string) {
	fileName := prefix + strconv.FormatInt(time.Now().Unix(), 10)
	return fileName, uploadEmptyFile(fileName)
}

func uploadEmptyFile(key string) string {
	region := MustHaveEnv("AWS_TEST_BUCKET_REGION")
	bucket := MustHaveEnv("AWS_TEST_BUCKET_NAME")

	outputBuffer := new(bytes.Buffer)
	errorBuffer := new(bytes.Buffer)

	awsCmd := exec.Command("aws",
		"--region", region,
		"s3api",
		"put-object",
		"--bucket", bucket,
		"--key", key)
	awsCmd.Stdout = outputBuffer
	awsCmd.Stderr = errorBuffer

	err := awsCmd.Run()
	Expect(err).ToNot(HaveOccurred(), errorBuffer.String())

	var response PutResponse
	json.Unmarshal(outputBuffer.Bytes(), &response)

	return response.VersionId
}

func deleteFile(key string) string {
	region := MustHaveEnv("AWS_TEST_BUCKET_REGION")
	bucket := MustHaveEnv("AWS_TEST_BUCKET_NAME")

	outputBuffer := new(bytes.Buffer)

	awsCmd := exec.Command("aws",
		"--region", region,
		"s3api",
		"delete-object",
		"--bucket", bucket,
		"--key", key)
	awsCmd.Stdout = outputBuffer
	err := awsCmd.Run()
	Expect(err).ToNot(HaveOccurred())

	var response PutResponse
	json.Unmarshal(outputBuffer.Bytes(), &response)

	return response.VersionId
}

func deleteVersion(key, versionId string) {
	region := MustHaveEnv("AWS_TEST_BUCKET_REGION")
	bucket := MustHaveEnv("AWS_TEST_BUCKET_NAME")

	fmt.Println("aws",
		"--region", region,
		"s3api",
		"delete-object",
		"--bucket", bucket,
		"--key", key,
		"--version-id", versionId)
	err := exec.Command("aws",
		"--region", region,
		"s3api",
		"delete-object",
		"--bucket", bucket,
		"--key", key,
		"--version-id", versionId).Run()
	Expect(err).NotTo(HaveOccurred())
}

type PutResponse struct {
	VersionId string
}
