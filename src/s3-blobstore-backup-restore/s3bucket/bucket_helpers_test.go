package s3bucket_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

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

func listFiles(bucket, endpoint string) []string {
	baseCmd := constructBaseCmd(endpoint)
	baseCmd = append(baseCmd, "s3api",
		"list-objects",
		"--bucket", bucket)

	outputBuffer := runAwsCommand(baseCmd)

	var response ListResponse
	_ = json.Unmarshal(outputBuffer.Bytes(), &response)

	var keys []string
	for _, entry := range response.Contents {
		keys = append(keys, entry.Key)
	}

	return keys
}

func getFileContents(bucket, endpoint, key string) string {
	baseCmd := constructBaseCmd(endpoint)
	baseCmd = append(baseCmd, "s3",
		"cp",
		fmt.Sprintf("s3://%s/%s", bucket, key),
		"-")

	outputBuffer := runAwsCommand(baseCmd)

	return outputBuffer.String()
}

func uploadFile(bucket, endpoint, key, body string) string {
	bodyFile, _ := os.CreateTemp("", "")
	_, _ = bodyFile.WriteString(body)
	bodyFile.Close()

	baseCmd := constructBaseCmd(endpoint)
	baseCmd = append(baseCmd, "s3api",
		"put-object",
		"--bucket", bucket,
		"--key", key,
		"--body", bodyFile.Name())

	outputBuffer := runAwsCommand(baseCmd)

	var response PutResponse
	_ = json.Unmarshal(outputBuffer.Bytes(), &response)

	return response.VersionId
}

func downloadFileToTmp(bucket, endpoint, key string) string {
	bodyFile, _ := os.CreateTemp("", "")
	bodyFile.Close()

	baseCmd := constructBaseCmd(endpoint)
	baseCmd = append(baseCmd, "s3api",
		"get-object",
		"--bucket", bucket,
		"--key", key,
		bodyFile.Name())

	runAwsCommandWithTimeout(baseCmd, 15*time.Minute)

	return bodyFile.Name()
}

func setUpUnversionedBucket(region, endpoint string) string {
	bucketName := "sdk-integration-test-" + strconv.FormatInt(time.Now().UnixNano(), 10)

	baseCmd := constructBaseCmd(endpoint)
	baseCmd = append(baseCmd, "s3api",
		"create-bucket",
		"--bucket", bucketName,
		"--region", region,
		"--create-bucket-configuration", "LocationConstraint="+region)

	runAwsCommand(baseCmd)
	return bucketName
}

func setUpVersionedBucket(region, endpoint string) string {
	testBucketName := setUpUnversionedBucket(region, endpoint)
	enableBucketVersioning(testBucketName, endpoint)
	return testBucketName
}

func enableBucketVersioning(bucket, endpoint string) {
	setBucketVersioning("Status=Enabled", bucket, endpoint)
}

func disableBucketVersioning(bucket, endpoint string) {
	setBucketVersioning("Status=Suspended", bucket, endpoint)
}

func setBucketVersioning(versioningConfig, bucket, endpoint string) {
	baseCmd := constructBaseCmd(endpoint)
	baseCmd = append(baseCmd, "s3api",
		"put-bucket-versioning",
		"--bucket", bucket,
		"--versioning-configuration", versioningConfig)

	runAwsCommand(baseCmd)
}

func tearDownBucket(bucket, endpoint string) {
	baseCmd := constructBaseCmd(endpoint)
	baseCmd = append(baseCmd, "s3", "rb", "s3://"+bucket, "--force")

	runAwsCommand(baseCmd)
}

func tearDownVersionedBucket(bucket, endpoint string) {
	clearOutVersionedBucket(bucket, endpoint)
	tearDownBucket(bucket, endpoint)
}

func deleteFile(bucket, endpoint, key string) string {
	baseCmd := constructBaseCmd(endpoint)
	baseCmd = append(baseCmd, "s3api",
		"delete-object",
		"--bucket", bucket,
		"--key", key)

	outputBuffer := runAwsCommand(baseCmd)

	var response PutResponse
	_ = json.Unmarshal(outputBuffer.Bytes(), &response)

	return response.VersionId
}

func deleteVersion(bucket, endpoint, key, versionId string) {
	baseCmd := constructBaseCmd(endpoint)
	baseCmd = append(baseCmd, "s3api",
		"delete-object",
		"--bucket", bucket,
		"--key", key,
		"--version-id", versionId)
	runAwsCommand(baseCmd)
}

func clearOutVersionedBucket(bucket, endpoint string) {
	baseCmd := constructBaseCmd(endpoint)
	baseCmd = append(baseCmd, "s3api",
		"list-object-versions",
		"--bucket", bucket)

	outputBuffer := runAwsCommand(baseCmd)

	var response VersionsResponse
	_ = json.Unmarshal(outputBuffer.Bytes(), &response)

	for _, version := range response.Versions {
		deleteVersion(bucket, endpoint, version.Key, version.VersionId)
	}

	for _, version := range response.DeleteMarkers {
		deleteVersion(bucket, endpoint, version.Key, version.VersionId)
	}
}

func constructBaseCmd(endpoint string) []string {
	if endpoint != "" {
		return []string{"--endpoint", endpoint}
	}
	return []string{}
}

func newAwsCommand(baseCmd []string) *exec.Cmd {
	commandArguments := append([]string{"--profile", AssumedRoleProfileName}, baseCmd...)
	awsCmd := exec.Command("aws", commandArguments...)
	awsCmd.Env = append(awsCmd.Env, "AWS_ACCESS_KEY_ID="+AccessKey)
	awsCmd.Env = append(awsCmd.Env, "AWS_SECRET_ACCESS_KEY="+SecretKey)

	return awsCmd
}

func runAwsCommand(baseCmd []string) *bytes.Buffer {
	return runAwsCommandWithTimeout(baseCmd, 1*time.Minute)
}

func runAwsCommandWithTimeout(baseCmd []string, timeout time.Duration) *bytes.Buffer {
	outputBuffer := new(bytes.Buffer)
	awsCmd := newAwsCommand(baseCmd)

	_, _ = fmt.Fprintf(GinkgoWriter, "Running command: aws %s", strings.Join(baseCmd, " "))
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
