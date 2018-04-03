package blobstore_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os/exec"
	"strings"
	"time"

	. "github.com/cloudfoundry-incubator/blobstore-backup-restore"

	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

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

type TestS3Bucket struct {
	Name   string
	Region string
}

func listFiles(bucket, endpoint string, creds S3AccessKey) []string {
	baseCmd := constructBaseCmd(endpoint)
	baseCmd = append(baseCmd, "s3api",
		"list-objects",
		"--bucket", bucket)

	outputBuffer := runAwsCommand(creds.Id, creds.Secret, baseCmd)

	var response ListResponse
	json.Unmarshal(outputBuffer.Bytes(), &response)

	var keys []string
	for _, entry := range response.Contents {
		keys = append(keys, entry.Key)
	}

	return keys
}

func getFileContents(bucket, endpoint, key string, creds S3AccessKey) string {
	baseCmd := constructBaseCmd(endpoint)
	baseCmd = append(baseCmd, "s3",
		"cp",
		fmt.Sprintf("s3://%s/%s", bucket, key),
		"-")

	outputBuffer := runAwsCommand(creds.Id, creds.Secret, baseCmd)

	return outputBuffer.String()
}

func uploadFile(bucket, endpoint, key, body string, creds S3AccessKey) string {
	bodyFile, _ := ioutil.TempFile("", "")
	bodyFile.WriteString(body)
	bodyFile.Close()

	baseCmd := constructBaseCmd(endpoint)
	baseCmd = append(baseCmd, "s3api",
		"put-object",
		"--bucket", bucket,
		"--key", key,
		"--body", bodyFile.Name())

	outputBuffer := runAwsCommand(creds.Id, creds.Secret, baseCmd)

	var response PutResponse
	json.Unmarshal(outputBuffer.Bytes(), &response)

	return response.VersionId
}

func downloadFileToTmp(bucket, endpoint, key string, creds S3AccessKey) string {
	bodyFile, _ := ioutil.TempFile("", "")
	bodyFile.Close()

	baseCmd := constructBaseCmd(endpoint)
	baseCmd = append(baseCmd, "s3api",
		"get-object",
		"--bucket", bucket,
		"--key", key,
		bodyFile.Name())

	runAwsCommandWithTimout(creds.Id, creds.Secret, baseCmd, 5*time.Minute)

	return bodyFile.Name()
}

func setUpS3UnversionedBucket(region, endpoint string, creds S3AccessKey) TestS3Bucket {
	bucketName := "sdk-integration-test-" + strconv.FormatInt(time.Now().UnixNano(), 10)

	baseCmd := constructBaseCmd(endpoint)
	baseCmd = append(baseCmd, "s3api",
		"create-bucket",
		"--bucket", bucketName,
		"--region", region,
		"--create-bucket-configuration", "LocationConstraint="+region)

	runAwsCommand(creds.Id, creds.Secret, baseCmd)
	return TestS3Bucket{Name: bucketName, Region: region}
}

func setUpVersionedS3Bucket(region, endpoint string, creds S3AccessKey) TestS3Bucket {
	testBucket := setUpS3UnversionedBucket(region, endpoint, creds)
	enableBucketVersioning(testBucket.Name, endpoint, creds)
	return testBucket
}

func enableBucketVersioning(bucket, endpoint string, creds S3AccessKey) {
	baseCmd := constructBaseCmd(endpoint)
	baseCmd = append(baseCmd, "s3api",
		"put-bucket-versioning",
		"--bucket", bucket,
		"--versioning-configuration", "Status=Enabled")

	runAwsCommand(creds.Id, creds.Secret, baseCmd)
}

func tearDownBucket(bucket, endpoint string, creds S3AccessKey) {
	baseCmd := constructBaseCmd(endpoint)
	baseCmd = append(baseCmd, "s3", "rb", "s3://"+bucket, "--force")

	runAwsCommand(creds.Id, creds.Secret, baseCmd)
}

func tearDownVersionedBucket(bucket, endpoint string, creds S3AccessKey) {
	clearOutVersionedBucket(bucket, endpoint, creds)
	tearDownBucket(bucket, endpoint, creds)
}

func deleteFile(bucket, endpoint, key string, creds S3AccessKey) string {
	baseCmd := constructBaseCmd(endpoint)
	baseCmd = append(baseCmd, "s3api",
		"delete-object",
		"--bucket", bucket,
		"--key", key)

	outputBuffer := runAwsCommand(creds.Id, creds.Secret, baseCmd)

	var response PutResponse
	json.Unmarshal(outputBuffer.Bytes(), &response)

	return response.VersionId
}

func deleteVersion(bucket, endpoint, key, versionId string, creds S3AccessKey) {
	baseCmd := constructBaseCmd(endpoint)
	baseCmd = append(baseCmd, "s3api",
		"delete-object",
		"--bucket", bucket,
		"--key", key,
		"--version-id", versionId)
	runAwsCommand(creds.Id, creds.Secret, baseCmd)
}

func clearOutVersionedBucket(bucket, endpoint string, creds S3AccessKey) {
	baseCmd := constructBaseCmd(endpoint)
	baseCmd = append(baseCmd, "s3api",
		"list-object-versions",
		"--bucket", bucket)

	outputBuffer := runAwsCommand(creds.Id, creds.Secret, baseCmd)

	var response VersionsResponse
	json.Unmarshal(outputBuffer.Bytes(), &response)

	for _, version := range response.Versions {
		deleteVersion(bucket, endpoint, version.Key, version.VersionId, creds)
	}

	for _, version := range response.DeleteMarkers {
		deleteVersion(bucket, endpoint, version.Key, version.VersionId, creds)
	}
}

func constructBaseCmd(endpoint string) []string {
	if endpoint != "" {
		return []string{"--endpoint", endpoint}
	}
	return []string{}
}

func newAwsCommand(accessKey string, secretKey string, baseCmd []string) *exec.Cmd {
	awsCmd := exec.Command("aws", baseCmd...)
	awsCmd.Env = append(awsCmd.Env, "AWS_ACCESS_KEY_ID="+accessKey)
	awsCmd.Env = append(awsCmd.Env, "AWS_SECRET_ACCESS_KEY="+secretKey)

	return awsCmd
}

func runAwsCommand(accessKey string, secretKey string, baseCmd []string) *bytes.Buffer {
	return runAwsCommandWithTimout(accessKey, secretKey, baseCmd, 1*time.Minute)
}

func runAwsCommandWithTimout(accessKey string, secretKey string, baseCmd []string, timeout time.Duration) *bytes.Buffer {
	outputBuffer := new(bytes.Buffer)
	awsCmd := newAwsCommand(accessKey, secretKey, baseCmd)

	fmt.Fprintf(GinkgoWriter, "Running command: aws %s", strings.Join(baseCmd, " "))
	session, err := gexec.Start(awsCmd, io.MultiWriter(GinkgoWriter, outputBuffer), GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session, timeout).Should(gexec.Exit())
	Expect(session.ExitCode()).To(BeZero(), string(session.Err.Contents()))

	return outputBuffer
}

func shasum(filePath string) string {
	output, err := exec.Command("shasum", filePath).Output()
	Expect(err).NotTo(HaveOccurred())
	md5 := strings.Split(string(output), " ")[0]
	return md5
}
