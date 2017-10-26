package blobstore_test

import (
	. "github.com/cloudfoundry-incubator/blobstore-backup-restore"

	"os"

	"os/exec"

	"bytes"
	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("S3Bucket", func() {
	var bucket S3Bucket
	var creds S3AccessKey

	var region = "eu-west-1"
	var bucketName = "bbr-integration-test-bucket"

	Describe("Versions", func() {
		var versions []Version
		var err error

		var firstVersionOfTest1 string
		var secondVersionOfTest1 string
		var thirdVersionOfTest1 string
		var firstVersionOfTest2 string
		var deletedVersionOfTest2 string

		JustBeforeEach(func() {
			bucket = NewS3Bucket("aws", bucketName, region, creds)
			versions, err = bucket.Versions()
		})

		Context("when retrieving versions succeeds", func() {
			BeforeEach(func() {
				creds = S3AccessKey{
					Id:     os.Getenv("AWS_ACCESS_KEY_ID"),
					Secret: os.Getenv("AWS_SECRET_ACCESS_KEY"),
				}

				firstVersionOfTest1 = uploadEmptyFile(region, bucketName, "test-1")
				secondVersionOfTest1 = uploadEmptyFile(region, bucketName, "test-1")
				thirdVersionOfTest1 = uploadEmptyFile(region, bucketName, "test-1")
				firstVersionOfTest2 = uploadEmptyFile(region, bucketName, "test-2")
				deletedVersionOfTest2 = deleteFile(region, bucketName, "test-2")
			})

			AfterEach(func() {
				deleteVersion(region, bucketName, "test-1", firstVersionOfTest1)
				deleteVersion(region, bucketName, "test-1", secondVersionOfTest1)
				deleteVersion(region, bucketName, "test-1", thirdVersionOfTest1)
				deleteVersion(region, bucketName, "test-2", firstVersionOfTest2)
				deleteVersion(region, bucketName, "test-2", deletedVersionOfTest2)
			})

			It("returns a list of all versions in the bucket", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(versions).To(ConsistOf(
					Version{Id: firstVersionOfTest1, Key: "test-1", IsLatest: false},
					Version{Id: secondVersionOfTest1, Key: "test-1", IsLatest: false},
					Version{Id: thirdVersionOfTest1, Key: "test-1", IsLatest: true},
					Version{Id: firstVersionOfTest2, Key: "test-2", IsLatest: false},
				))
			})
		})

		Context("when retrieving versions fails", func() {
			BeforeEach(func() {
				creds = S3AccessKey{}
			})

			It("returns the error", func() {
				Expect(versions).To(BeNil())
				Expect(err).To(MatchError(ContainSubstring("An error occurred")))
			})
		})
	})
})

func uploadEmptyFile(region, bucket, key string) string {
	outputBuffer := new(bytes.Buffer)

	awsCmd := exec.Command("aws", "--region", region, "s3api", "put-object", "--bucket", bucket, "--key", key)
	awsCmd.Stdout = outputBuffer
	err := awsCmd.Run()
	Expect(err).ToNot(HaveOccurred())

	var response PutResponse
	json.Unmarshal(outputBuffer.Bytes(), &response)

	return response.VersionId
}

func deleteFile(region, bucket, key string) string {
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

func deleteVersion(region, bucket, key, versionId string) {
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
