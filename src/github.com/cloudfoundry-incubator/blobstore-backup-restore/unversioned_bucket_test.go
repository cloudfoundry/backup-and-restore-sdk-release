package blobstore_test

import (
	"os"

	. "github.com/cloudfoundry-incubator/blobstore-backup-restore"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

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

	AfterEach(func() {
		tearDownBucket(liveBucket.Name, endpoint, creds)
	})

	Describe("ListFiles", func() {
		var files []string

		BeforeEach(func() {
			bucketObjectUnderTest, err = NewS3UnversionedBucket(liveBucket.Name, liveBucket.Region, endpoint, creds)
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
