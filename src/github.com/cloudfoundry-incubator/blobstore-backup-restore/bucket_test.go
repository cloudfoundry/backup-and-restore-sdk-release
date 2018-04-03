package blobstore_test

import (
	. "github.com/cloudfoundry-incubator/blobstore-backup-restore"

	"os"

	"os/exec"

	"bytes"
	"encoding/json"

	"io/ioutil"

	"fmt"

	"strconv"
	"time"

	"strings"

	"io"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("S3VersionedBucket", func() {
	var bucketObjectUnderTest S3VersionedBucket
	var err error

	var firstVersionOfFile1 string
	var secondVersionOfFile1 string
	var thirdVersionOfFile1 string
	var firstVersionOfFile2 string
	var deletedVersionOfFile2 string

	RunVersionedBucketTests := func(mainRegion, secondaryRegion, endpoint, accessKey, secretKey string) {
		var bucket TestS3Bucket

		var creds = S3AccessKey{
			Id:     accessKey,
			Secret: secretKey,
		}

		BeforeEach(func() {
			bucket = setUpVersionedS3Bucket(mainRegion, endpoint, creds)

			firstVersionOfFile1 = uploadFile(bucket.Name, endpoint, "test-1", "1-A", creds)
			secondVersionOfFile1 = uploadFile(bucket.Name, endpoint, "test-1", "1-B", creds)
			thirdVersionOfFile1 = uploadFile(bucket.Name, endpoint, "test-1", "1-C", creds)
			firstVersionOfFile2 = uploadFile(bucket.Name, endpoint, "test-2", "2-A", creds)
			deletedVersionOfFile2 = deleteFile(bucket.Name, endpoint, "test-2", creds)

			bucketObjectUnderTest, err = NewS3VersionedBucket(bucket.Name, bucket.Region, endpoint, creds)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			tearDownVersionedBucket(bucket.Name, endpoint, creds)
		})

		Describe("Versions", func() {
			var versions []Version

			JustBeforeEach(func() {
				versions, err = bucketObjectUnderTest.Versions()
			})

			Context("when retrieving versions succeeds", func() {
				It("returns a list of all versions in the bucket", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(versions).To(ConsistOf(
						Version{Id: firstVersionOfFile1, Key: "test-1", IsLatest: false},
						Version{Id: secondVersionOfFile1, Key: "test-1", IsLatest: false},
						Version{Id: thirdVersionOfFile1, Key: "test-1", IsLatest: true},
						Version{Id: firstVersionOfFile2, Key: "test-2", IsLatest: false},
					))
				})
			})

			Context("when the bucket can't be reached", func() {
				BeforeEach(func() {
					bucketObjectUnderTest, err = NewS3VersionedBucket(
						bucket.Name,
						bucket.Region,
						endpoint,
						S3AccessKey{Id: "NOT RIGHT", Secret: "NOT RIGHT"},
					)
					Expect(err).NotTo(HaveOccurred())
				})

				It("returns the error", func() {
					Expect(versions).To(BeNil())
					Expect(err).To(MatchError(MatchRegexp("could not check if bucket (.*) is versioned")))
				})
			})

			Context("when the bucket is not versioned", func() {

				var unversionedBucket TestS3Bucket

				BeforeEach(func() {
					unversionedBucket = setUpS3UnversionedBucket(mainRegion, endpoint, creds)
					uploadFile(unversionedBucket.Name, endpoint, "unversioned-test", "UNVERSIONED-TEST", creds)

					bucketObjectUnderTest, err = NewS3VersionedBucket(unversionedBucket.Name, unversionedBucket.Region, endpoint, creds)
				})

				AfterEach(func() {
					tearDownBucket(unversionedBucket.Name, endpoint, creds)
				})

				It("fails", func() {
					Expect(versions).To(BeNil())
					Expect(err).To(MatchError(ContainSubstring("is not versioned")))
				})
			})

			Context("when the bucket has a lot of files", func() {
				BeforeEach(func() {
					bucketObjectUnderTest, err = NewS3VersionedBucket("sdk-big-bucket-integration-test",
						mainRegion, endpoint, creds)
				})

				It("works", func() {
					versions, err := bucketObjectUnderTest.Versions()

					Expect(err).NotTo(HaveOccurred())
					Expect(len(versions)).To(Equal(2001))
				})
			})
		})

		Describe("CopyVersions from same bucket", func() {
			BeforeEach(func() {
				uploadFile(bucket.Name, endpoint, "test-3", "TEST-3-A", creds)
			})

			JustBeforeEach(func() {
				err = bucketObjectUnderTest.CopyVersions(bucket.Region, bucket.Name, []BlobVersion{
					{BlobKey: "test-1", Id: secondVersionOfFile1},
					{BlobKey: "test-2", Id: firstVersionOfFile2},
				})
			})

			Context("when putting versions succeeds", func() {
				It("restores files to versions specified in the backup and does not delete pre-existing blobs", func() {
					Expect(err).NotTo(HaveOccurred())

					Expect(listFiles(bucket.Name, endpoint, creds)).To(ConsistOf(
						"test-1", "test-2", "test-3"))
					Expect(getFileContents(bucket.Name, endpoint, "test-1", creds)).To(Equal(
						"1-B"))
					Expect(getFileContents(bucket.Name, endpoint, "test-2", creds)).To(Equal(
						"2-A"))
					Expect(getFileContents(bucket.Name, endpoint, "test-3", creds)).To(Equal(
						"TEST-3-A"))
				})
			})

			Context("when putting versions fails", func() {
				BeforeEach(func() {
					bucketObjectUnderTest, err = NewS3VersionedBucket(bucket.Name,
						bucket.Region, endpoint, S3AccessKey{})
				})

				It("errors", func() {
					Expect(err).To(HaveOccurred())
				})
			})

			Context("when the bucket is not versioned", func() {

				var unversionedBucket TestS3Bucket

				BeforeEach(func() {
					unversionedBucket = setUpS3UnversionedBucket(mainRegion, endpoint, creds)
					uploadFile(unversionedBucket.Name, endpoint, "unversioned-test", "UNVERSIONED-TEST", creds)

					bucketObjectUnderTest, err = NewS3VersionedBucket(unversionedBucket.Name, unversionedBucket.Region, endpoint, creds)
				})

				AfterEach(func() {
					tearDownBucket(unversionedBucket.Name, endpoint, creds)
				})

				It("fails", func() {
					Expect(err).To(MatchError(ContainSubstring("is not versioned")))
				})
			})
		})

		Describe("CopyVersions from different bucket in different region", func() {
			var secondaryBucket TestS3Bucket
			var versionOfFileWhichWasSubsequentlyDeleted, versionOfFileToBeRestored string

			BeforeEach(func() {
				clearOutVersionedBucket(bucket.Name, endpoint, creds)
				secondaryBucket = setUpVersionedS3Bucket(secondaryRegion, endpoint, creds)
				versionOfFileToBeRestored = uploadFile(
					secondaryBucket.Name,
					endpoint,
					"file-to-restore",
					"whatever",
					creds,
				)
				versionOfFileWhichWasSubsequentlyDeleted = uploadFile(
					secondaryBucket.Name,
					endpoint,
					"deleted-file-to-restore",
					"whatever",
					creds,
				)
				deleteFile(secondaryBucket.Name, endpoint, "deleted-file-to-restore", creds)
				uploadFile(bucket.Name, endpoint, "file-to-be-destroyed-by-restore",
					"whatever", creds)
			})

			JustBeforeEach(func() {
				err = bucketObjectUnderTest.CopyVersions(secondaryBucket.Region, secondaryBucket.Name,
					[]BlobVersion{
						{BlobKey: "file-to-restore", Id: versionOfFileToBeRestored},
						{BlobKey: "deleted-file-to-restore", Id: versionOfFileWhichWasSubsequentlyDeleted},
					})
			})

			It("restores files from the secondary to the main bucket and does not delete pre-existing blobs", func() {
				Expect(err).NotTo(HaveOccurred())

				Expect(listFiles(bucket.Name, endpoint, creds)).To(
					ConsistOf("file-to-restore", "deleted-file-to-restore", "file-to-be-destroyed-by-restore"),
				)
			})

			AfterEach(func() {
				tearDownVersionedBucket(secondaryBucket.Name, endpoint, creds)
			})
		})
	}

	Describe("AWS S3 bucket", func() {
		RunVersionedBucketTests(
			"eu-west-1",
			"us-west-1",
			"",
			os.Getenv("TEST_AWS_ACCESS_KEY_ID"),
			os.Getenv("TEST_AWS_SECRET_ACCESS_KEY"),
		)
	})

	Describe("ECS S3-compatible bucket", func() {
		RunVersionedBucketTests(
			"eu-west-1",
			"us-east-1",
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

			creds := S3AccessKey{
				Id:     os.Getenv("TEST_AWS_ACCESS_KEY_ID"),
				Secret: os.Getenv("TEST_AWS_SECRET_ACCESS_KEY"),
			}

			clearOutVersionedBucket(bucketName, endpoint, creds)
			bucketObjectUnderTest, err = NewS3VersionedBucket(bucketName, region, endpoint, creds)
		})

		Context("when backup an empty bucket", func() {
			It("does not fail", func() {
				_, err := bucketObjectUnderTest.Versions()
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when restore from an empty bucket", func() {
			It("does not fail", func() {
				err := bucketObjectUnderTest.CopyVersions(region, bucketName, []BlobVersion{})
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("CopyVersions with a big file on AWS", func() {
		var destinationBucket TestS3Bucket
		var region string
		var endpoint string
		var creds S3AccessKey

		BeforeEach(func() {
			region = "eu-west-1"
			endpoint = ""

			creds = S3AccessKey{
				Id:     os.Getenv("TEST_AWS_ACCESS_KEY_ID"),
				Secret: os.Getenv("TEST_AWS_SECRET_ACCESS_KEY"),
			}

			destinationBucket = setUpVersionedS3Bucket(region, endpoint, creds)

			bucketObjectUnderTest, err = NewS3VersionedBucket(destinationBucket.Name, region, endpoint, creds)
		})

		AfterEach(func() {
			clearOutVersionedBucket(destinationBucket.Name, endpoint, creds)
			tearDownBucket(destinationBucket.Name, endpoint, creds)
		})

		It("works", func() {
			err := bucketObjectUnderTest.CopyVersions(
				"eu-west-1",
				"large-blob-test-bucket", []BlobVersion{
					{BlobKey: "big_file", Id: "YfWcz5KoJzfjKB9gnBI6q7ue_jZGTvkw"},
				})

			Expect(err).NotTo(HaveOccurred())
			Expect(listFiles(destinationBucket.Name, endpoint, creds)).To(ConsistOf("big_file"))

			localFilePath := downloadFileToTmp(destinationBucket.Name, endpoint, "big_file", creds)
			Expect(shasum(localFilePath)).To(Equal("188f500de28479d67e7375566750472e58e4cec1"))
		})
	})
})

var _ = Describe("S3UnversionedBucket", func() {
	var (
		liveBucket            TestS3Bucket
		bucketObjectUnderTest S3UnversionedBucket
		err                   error
		liveRegion            string
		endpoint              string
		accessKey             string
		secretKey             string
		testFile1             string
		testFile2             string
		creds                 S3AccessKey
	)

	BeforeEach(func() {
		liveRegion = "eu-west-1"
		endpoint = ""
		accessKey = os.Getenv("TEST_AWS_ACCESS_KEY_ID")
		secretKey = os.Getenv("TEST_AWS_SECRET_ACCESS_KEY")
		creds = S3AccessKey{Id: accessKey, Secret: secretKey}

		liveBucket = setUpS3UnversionedBucket(liveRegion, endpoint, creds)
		testFile1 = uploadFile(liveBucket.Name, endpoint, "path1/file1", "FILE1", creds)
		testFile2 = uploadFile(liveBucket.Name, endpoint, "path2/file2", "FILE2", creds)

	})

	Describe("ListFiles", func() {
		var files []string

		BeforeEach(func() {
			bucketObjectUnderTest, err = NewS3UnversionedBucket(liveBucket.Name, liveBucket.Region, endpoint, creds)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			tearDownBucket(liveBucket.Name, endpoint, creds)
		})

		JustBeforeEach(func() {
			files, err = bucketObjectUnderTest.ListFiles()
		})

		It("should list all the files", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(files).To(ConsistOf([]string{"path1/file1", "path2/file2"}))
		})

		Context("when s3 list-objects errors", func() {
			BeforeEach(func() {
				bucketObjectUnderTest, err = NewS3UnversionedBucket("does-not-exist", "eu-west-1", "", creds)
			})
			It("errors", func() {
				Expect(err).To(MatchError(ContainSubstring("failed to list files from bucket does-not-exist")))
			})

		})

		Context("when the bucket has a lot of files", func() {
			BeforeEach(func() {
				bucketObjectUnderTest, err = NewS3UnversionedBucket(
					"sdk-unversioned-big-bucket-integration-test", "eu-west-1", endpoint, creds)
			})

			It("works", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(len(files)).To(Equal(2001))
			})
		})
	})

	Describe("Copy", func() {
		var (
			backedUpFiles         []string
			bucketObjectUnderTest S3UnversionedBucket
			backupBucket          TestS3Bucket
			backupRegion          string
		)

		BeforeEach(func() {
			backupRegion = "us-west-1"
			backupBucket = setUpS3UnversionedBucket(backupRegion, endpoint, creds)

			bucketObjectUnderTest, err = NewS3UnversionedBucket(backupBucket.Name, backupBucket.Region, endpoint, creds)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			tearDownBucket(backupBucket.Name, endpoint, creds)
		})

		JustBeforeEach(func() {
			err = bucketObjectUnderTest.Copy(
				"path1/file1", "2012_02_13_23_12_02/bucketIdFromDeployment",
				liveBucket.Name, liveBucket.Region)
		})

		It("copies the file to the backup bucket", func() {
			Expect(err).NotTo(HaveOccurred())
			backedUpFiles = listFiles(bucketObjectUnderTest.Name(), endpoint, creds)
			Expect(err).NotTo(HaveOccurred())
			Expect(backedUpFiles).To(ConsistOf([]string{"2012_02_13_23_12_02/bucketIdFromDeployment/path1/file1"}))
			Expect(getFileContents(
				backupBucket.Name,
				endpoint,
				"2012_02_13_23_12_02/bucketIdFromDeployment/path1/file1",
				creds),
			).To(Equal(
				"FILE1"))
		})
	})

	Describe("CopyFiles with a big file on AWS", func() {
		var endpoint string
		var creds S3AccessKey
		var preExistingBigFileBucketConfig TestS3Bucket
		var destinationBucket TestS3Bucket
		var err error

		BeforeEach(func() {
			endpoint = ""
			creds = S3AccessKey{
				Id:     os.Getenv("TEST_AWS_ACCESS_KEY_ID"),
				Secret: os.Getenv("TEST_AWS_SECRET_ACCESS_KEY"),
			}
			preExistingBigFileBucketConfig = TestS3Bucket{
				Name:   "large-blob-test-bucket-unversioned",
				Region: "eu-west-1",
			}

			destinationBucket = setUpS3UnversionedBucket("eu-west-1", endpoint, creds)

			bucketObjectUnderTest, err = NewS3UnversionedBucket(destinationBucket.Name, destinationBucket.Region, endpoint, creds)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			tearDownBucket(destinationBucket.Name, endpoint, creds)
		})

		JustBeforeEach(func() {
			err = bucketObjectUnderTest.Copy("big_file", "path/to/file",
				preExistingBigFileBucketConfig.Name, preExistingBigFileBucketConfig.Region)

		})
		It("works", func() {
			By("succeeding")
			Expect(err).NotTo(HaveOccurred())

			By("copying the large file")
			Expect(listFiles(destinationBucket.Name, endpoint, creds)).To(ConsistOf("path/to/file/big_file"))

			By("not corrupting the large file")
			Expect(
				shasum(downloadFileToTmp(destinationBucket.Name, endpoint, "path/to/file/big_file", creds))).To(
				Equal("188f500de28479d67e7375566750472e58e4cec1"))
		})
	})

})

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
