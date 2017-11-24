package blobstore_test

import (
	. "github.com/cloudfoundry-incubator/blobstore-backup-restore"

	"os"

	"os/exec"

	"bytes"
	"encoding/json"

	"io/ioutil"

	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("S3Bucket", func() {
	var bucket S3Bucket
	var creds S3AccessKey

	var firstVersionOfTest1 string
	var secondVersionOfTest1 string
	var thirdVersionOfTest1 string
	var firstVersionOfTest2 string
	var deletedVersionOfTest2 string

	RunBucketTests := func(region, bucketName, endpoint, accessKey, secretKey string) {
		BeforeEach(func() {
			firstVersionOfTest1 = uploadFile(region, bucketName, endpoint, "test-1", "TEST-1-A", accessKey, secretKey)
			secondVersionOfTest1 = uploadFile(region, bucketName, endpoint, "test-1", "TEST-1-B", accessKey, secretKey)
			thirdVersionOfTest1 = uploadFile(region, bucketName, endpoint, "test-1", "TEST-1-C", accessKey, secretKey)
			firstVersionOfTest2 = uploadFile(region, bucketName, endpoint, "test-2", "TEST-2-A", accessKey, secretKey)
			deletedVersionOfTest2 = deleteFile(region, bucketName, endpoint, "test-2", accessKey, secretKey)

		})

		AfterEach(func() {
			deleteAllVersions(region, bucketName, endpoint, accessKey, secretKey)
		})

		JustBeforeEach(func() {
			bucket = NewS3Bucket("aws", bucketName, region, endpoint, creds)
		})

		Describe("Versions", func() {
			var versions []Version
			var err error

			JustBeforeEach(func() {
				versions, err = bucket.Versions()
			})

			Context("when retrieving versions succeeds", func() {
				BeforeEach(func() {
					creds = S3AccessKey{
						Id:     accessKey,
						Secret: secretKey,
					}
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

		Describe("PutVersions", func() {
			var err error

			BeforeEach(func() {
				uploadFile(region, bucketName, endpoint, "test-3", "TEST-3-A", accessKey, secretKey)
			})

			JustBeforeEach(func() {
				err = bucket.PutVersions(region, bucketName, []LatestVersion{
					{BlobKey: "test-1", Id: secondVersionOfTest1},
					{BlobKey: "test-2", Id: firstVersionOfTest2},
				})
			})

			Context("when putting versions succeeds", func() {
				BeforeEach(func() {
					creds = S3AccessKey{
						Id:     accessKey,
						Secret: secretKey,
					}
				})

				It("restores the versions to the specified ones", func() {
					Expect(err).NotTo(HaveOccurred())

					Expect(listFiles(region, bucketName, endpoint, accessKey, secretKey)).To(ConsistOf("test-1", "test-2"))
					Expect(getFileContents(region, bucketName, endpoint, "test-1", accessKey, secretKey)).To(Equal("TEST-1-B"))
					Expect(getFileContents(region, bucketName, endpoint, "test-2", accessKey, secretKey)).To(Equal("TEST-2-A"))
				})
			})

			Context("when putting versions fails", func() {
				BeforeEach(func() {
					creds = S3AccessKey{}
				})

				It("errors", func() {
					Expect(err).To(HaveOccurred())
				})
			})
		})
	}

	Describe("AWS S3 bucket", func() {
		RunBucketTests(
			"eu-west-1",
			"bbr-integration-test-bucket",
			"",
			os.Getenv("TEST_AWS_ACCESS_KEY_ID"),
			os.Getenv("TEST_AWS_SECRET_ACCESS_KEY"),
		)
	})

	Describe("ECS S3-compatible bucket", func() {
		RunBucketTests(
			"eu-west-1",
			"ecs-bbr-integration-test-bucket",
			"https://object.ecstestdrive.com",
			os.Getenv("TEST_ECS_ACCESS_KEY_ID"),
			os.Getenv("TEST_ECS_SECRET_ACCESS_KEY"),
		)
	})

	Describe("Empty AWS S3 bucket", func() {
		var region string
		var bucketName string
		var endpoint string

		BeforeEach(func() {
			region = "eu-west-1"
			bucketName = "bbr-integration-test-bucket-empty"
			endpoint = ""

			creds = S3AccessKey{
				Id:     os.Getenv("TEST_AWS_ACCESS_KEY_ID"),
				Secret: os.Getenv("TEST_AWS_SECRET_ACCESS_KEY"),
			}

			deleteAllVersions(region, bucketName, endpoint, creds.Id, creds.Secret)
			bucket = NewS3Bucket("aws", bucketName, region, endpoint, creds)
		})

		Context("when backup an empty bucket", func() {
			It("does not fail", func() {
				_, err := bucket.Versions()
				Expect(err).NotTo(HaveOccurred())
			})
		})
		Context("when restore from an empty bucket", func() {
			It("does not fail", func() {
				err := bucket.PutVersions(region, bucketName, []LatestVersion{})
				Expect(err).NotTo(HaveOccurred())

			})
		})
	})
})

func listFiles(region, bucket, endpoint string, accessKey string, secretKey string) []string {
	outputBuffer := new(bytes.Buffer)

	baseCmd := constructBaseCmd(region, endpoint)
	baseCmd = append(baseCmd, "s3api",
		"list-objects",
		"--bucket", bucket)

	awsCmd := newAwsCommand(accessKey, secretKey, baseCmd)
	awsCmd.Stdout = outputBuffer
	err := awsCmd.Run()
	Expect(err).ToNot(HaveOccurred())

	var response ListResponse
	json.Unmarshal(outputBuffer.Bytes(), &response)

	keys := []string{}
	for _, entry := range response.Contents {
		keys = append(keys, entry.Key)
	}

	return keys
}

func constructBaseCmd(region, endpoint string) []string {
	if endpoint != "" {
		return []string{"--region", region, "--endpoint", endpoint}
	}
	return []string{"--region", region}
}

func getFileContents(region, bucket, endpoint, key string, accessKey string, secretKey string) string {
	outputBuffer := new(bytes.Buffer)

	baseCmd := constructBaseCmd(region, endpoint)
	baseCmd = append(baseCmd, "s3",
		"cp",
		fmt.Sprintf("s3://%s/%s", bucket, key),
		"-")

	awsCmd := newAwsCommand(accessKey, secretKey, baseCmd)
	awsCmd.Stdout = outputBuffer

	err := awsCmd.Run()
	Expect(err).ToNot(HaveOccurred())

	return outputBuffer.String()
}

func uploadFile(region, bucket, endpoint, key, body, accessKey, secretKey string) string {
	bodyFile, _ := ioutil.TempFile("", key)
	bodyFile.WriteString(body)
	bodyFile.Close()

	outputBuffer := new(bytes.Buffer)

	baseCmd := constructBaseCmd(region, endpoint)
	baseCmd = append(baseCmd, "s3api",
		"put-object",
		"--bucket", bucket,
		"--key", key,
		"--body", bodyFile.Name())
	awsCmd := newAwsCommand(accessKey, secretKey, baseCmd)
	awsCmd.Stdout = outputBuffer

	err := awsCmd.Run()
	Expect(err).ToNot(HaveOccurred())

	var response PutResponse
	json.Unmarshal(outputBuffer.Bytes(), &response)

	return response.VersionId
}

func deleteFile(region, bucket, endpoint, key, accessKey, secretKey string) string {
	outputBuffer := new(bytes.Buffer)

	baseCmd := constructBaseCmd(region, endpoint)
	baseCmd = append(baseCmd, "s3api",
		"delete-object",
		"--bucket", bucket,
		"--key", key)
	awsCmd := newAwsCommand(accessKey, secretKey, baseCmd)

	awsCmd.Stdout = outputBuffer
	err := awsCmd.Run()
	Expect(err).ToNot(HaveOccurred())

	var response PutResponse
	json.Unmarshal(outputBuffer.Bytes(), &response)

	return response.VersionId
}

func deleteVersion(region, bucket, endpoint, key, versionId string, accessKey string, secretKey string) {

	baseCmd := constructBaseCmd(region, endpoint)
	baseCmd = append(baseCmd, "s3api",
		"delete-object",
		"--bucket", bucket,
		"--key", key,
		"--version-id", versionId)
	awsCmd := newAwsCommand(accessKey, secretKey, baseCmd)
	err := awsCmd.Run()
	Expect(err).ToNot(HaveOccurred())
}

func deleteAllVersions(region, bucket, endpoint string, accessKey string, secretKey string) {
	outputBuffer := new(bytes.Buffer)

	baseCmd := constructBaseCmd(region, endpoint)
	baseCmd = append(baseCmd, "s3api",
		"list-object-versions",
		"--bucket", bucket)
	awsCmd := newAwsCommand(accessKey, secretKey, baseCmd)

	awsCmd.Stdout = outputBuffer
	err := awsCmd.Run()
	Expect(err).ToNot(HaveOccurred())

	var response VersionsResponse
	json.Unmarshal(outputBuffer.Bytes(), &response)

	for _, version := range response.Versions {
		deleteVersion(region, bucket, endpoint, version.Key, version.VersionId, accessKey, secretKey)
	}

	for _, version := range response.DeleteMarkers {
		deleteVersion(region, bucket, endpoint, version.Key, version.VersionId, accessKey, secretKey)
	}

}

func newAwsCommand(accessKey string, secretKey string, baseCmd []string) *exec.Cmd {
	awsCmd := exec.Command("aws", baseCmd...)
	awsCmd.Env = append(awsCmd.Env, "AWS_ACCESS_KEY_ID="+accessKey)
	awsCmd.Env = append(awsCmd.Env, "AWS_SECRET_ACCESS_KEY="+secretKey)

	return awsCmd
}

type PutResponse struct {
	VersionId string
}

type VersionsResponse struct {
	Versions      []VersionItem
	DeleteMarkers []VersionItem
}

type VersionItem struct {
	Key       string
	VersionId string
}

type ListResponse struct {
	Contents []ListResponseEntry
}

type ListResponseEntry struct {
	Key string
}
