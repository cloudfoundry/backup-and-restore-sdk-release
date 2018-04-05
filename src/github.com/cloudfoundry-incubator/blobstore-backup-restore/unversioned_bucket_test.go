package blobstore_test

import (
	"os"

	"github.com/cloudfoundry-incubator/blobstore-backup-restore/s3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("S3UnversionedBucket", func() {
	Describe("AWS S3 buckets", func() {
		RunUnversionedBucketTests(
			"eu-west-1",
			"us-west-1",
			"",
			os.Getenv("TEST_AWS_ACCESS_KEY_ID"),
			os.Getenv("TEST_AWS_SECRET_ACCESS_KEY"),
		)
	})

	Describe("ECS S3-compatible buckets", func() {
		RunUnversionedBucketTests(
			"eu-west-1",
			"us-west-1",
			"https://object.ecstestdrive.com",
			os.Getenv("TEST_ECS_ACCESS_KEY_ID"),
			os.Getenv("TEST_ECS_SECRET_ACCESS_KEY"),
		)
	})

	Describe("Copy with a big file on AWS", func() {
		var endpoint string
		var creds s3.S3AccessKey
		var preExistingBigFileBucketName string
		var destinationBucketName string
		var region string
		var bucketObjectUnderTest s3.UnversionedBucket
		var err error

		BeforeEach(func() {
			endpoint = ""
			creds = s3.S3AccessKey{
				Id:     os.Getenv("TEST_AWS_ACCESS_KEY_ID"),
				Secret: os.Getenv("TEST_AWS_SECRET_ACCESS_KEY"),
			}
			region = "eu-west-1"
			preExistingBigFileBucketName = "large-blob-test-bucket-unversioned"
			destinationBucketName = setUpS3UnversionedBucket("eu-west-1", endpoint, creds)

			bucketObjectUnderTest, err = s3.NewBucket(destinationBucketName, region, endpoint, creds)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			tearDownBucket(destinationBucketName, endpoint, creds)
		})

		JustBeforeEach(func() {
			err = bucketObjectUnderTest.Copy("big_file", "path/to/file",
				preExistingBigFileBucketName, region)

		})

		It("works", func() {
			By("succeeding")
			Expect(err).NotTo(HaveOccurred())

			By("copying the large file")
			Expect(listFiles(destinationBucketName, endpoint, creds)).To(ConsistOf("path/to/file/big_file"))

			By("not corrupting the large file")
			Expect(
				shasum(downloadFileToTmp(destinationBucketName, endpoint, "path/to/file/big_file", creds))).To(
				Equal("188f500de28479d67e7375566750472e58e4cec1"))
		})
	})
})

func RunUnversionedBucketTests(liveRegion, backupRegion, endpoint, accessKey, secretKey string) {
	var (
		liveBucketName        string
		bucketObjectUnderTest s3.UnversionedBucket
		err                   error
		testFile1             string
		testFile2             string
		creds                 s3.S3AccessKey
	)

	BeforeEach(func() {
		creds = s3.S3AccessKey{Id: accessKey, Secret: secretKey}

		liveBucketName = setUpS3UnversionedBucket(liveRegion, endpoint, creds)
		testFile1 = uploadFile(liveBucketName, endpoint, "path1/file1", "FILE1", creds)
		testFile2 = uploadFile(liveBucketName, endpoint, "path2/file2", "FILE2", creds)
	})

	AfterEach(func() {
		tearDownBucket(liveBucketName, endpoint, creds)
	})

	Describe("ListFiles", func() {
		var files []string

		BeforeEach(func() {
			bucketObjectUnderTest, err = s3.NewBucket(liveBucketName, liveRegion, endpoint, creds)
			Expect(err).NotTo(HaveOccurred())
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
				bucketObjectUnderTest, err = s3.NewBucket("does-not-exist", liveRegion, endpoint, creds)
				Expect(err).NotTo(HaveOccurred())
			})

			It("errors", func() {
				Expect(err).To(MatchError(ContainSubstring("failed to list files from bucket does-not-exist")))
			})

		})

		Context("when the bucket has a lot of files", func() {
			BeforeEach(func() {
				bucketObjectUnderTest, err = s3.NewBucket("sdk-unversioned-big-bucket-integration-test", liveRegion, endpoint, creds)
				Expect(err).NotTo(HaveOccurred())
			})

			It("works", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(len(files)).To(Equal(2001))
			})
		})
	})

	Describe("Copy", func() {
		var (
			backedUpFiles    []string
			backupBucketName string
		)

		BeforeEach(func() {
			backupBucketName = setUpS3UnversionedBucket(backupRegion, endpoint, creds)

			bucketObjectUnderTest, err = s3.NewBucket(backupBucketName, backupRegion, endpoint, creds)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			tearDownBucket(backupBucketName, endpoint, creds)
		})

		JustBeforeEach(func() {
			err = bucketObjectUnderTest.Copy(
				"path1/file1", "2012_02_13_23_12_02/bucketIdFromDeployment",
				liveBucketName, liveRegion)
		})

		It("copies the file to the backup bucket", func() {
			Expect(err).NotTo(HaveOccurred())
			backedUpFiles = listFiles(bucketObjectUnderTest.Name(), endpoint, creds)
			Expect(err).NotTo(HaveOccurred())
			Expect(backedUpFiles).To(ConsistOf([]string{"2012_02_13_23_12_02/bucketIdFromDeployment/path1/file1"}))
			Expect(getFileContents(
				backupBucketName,
				endpoint,
				"2012_02_13_23_12_02/bucketIdFromDeployment/path1/file1",
				creds),
			).To(Equal("FILE1"))
		})
	})
}
