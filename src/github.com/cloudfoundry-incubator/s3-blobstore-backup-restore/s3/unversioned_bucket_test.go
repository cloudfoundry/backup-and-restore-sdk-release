package s3_test

import (
	"os"

	"github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/s3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UnversionedBucket", func() {
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

	Describe("CopyObject with a big file on AWS", func() {
		var endpoint string
		var creds s3.AccessKey
		var preExistingBigFileBucketName string
		var destinationBucketName string
		var region string
		var bucketObjectUnderTest s3.UnversionedBucket
		var err error

		BeforeEach(func() {
			endpoint = ""
			creds = s3.AccessKey{
				Id:     os.Getenv("TEST_AWS_ACCESS_KEY_ID"),
				Secret: os.Getenv("TEST_AWS_SECRET_ACCESS_KEY"),
			}
			region = "eu-west-1"
			preExistingBigFileBucketName = "large-blob-test-bucket-unversioned"
			destinationBucketName = setUpUnversionedBucket("eu-west-1", endpoint, creds)

			bucketObjectUnderTest, err = s3.NewBucket(destinationBucketName, region, endpoint, creds, false)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			tearDownBucket(destinationBucketName, endpoint, creds)
		})

		JustBeforeEach(func() {
			err = bucketObjectUnderTest.CopyObject(
				"big_file",
				"",
				"path/to/file",
				preExistingBigFileBucketName,
				region,
			)

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
		liveFile              string
		creds                 s3.AccessKey
	)

	BeforeEach(func() {
		creds = s3.AccessKey{Id: accessKey, Secret: secretKey}

		liveBucketName = setUpUnversionedBucket(liveRegion, endpoint, creds)
		testFile1 = uploadFile(liveBucketName, endpoint, "path1/file1", "FILE1", creds)
		liveFile = uploadFile(liveBucketName, endpoint, "live/location/leaf/node", "CONTENTS", creds)
		testFile2 = uploadFile(liveBucketName, endpoint, "path2/file2", "FILE2", creds)
	})

	AfterEach(func() {
		tearDownBucket(liveBucketName, endpoint, creds)
	})

	Describe("ListFiles", func() {
		var files []string

		BeforeEach(func() {
			bucketObjectUnderTest, err = s3.NewBucket(liveBucketName, liveRegion, endpoint, creds, false)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("When I ask for a list of all the files in the bucket", func() {

			JustBeforeEach(func() {
				files, err = bucketObjectUnderTest.ListFiles("")
			})

			It("should list all the files", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(files).To(ConsistOf([]string{"path1/file1", "live/location/leaf/node", "path2/file2"}))
			})

			Context("when s3 list-objects errors", func() {
				BeforeEach(func() {
					bucketObjectUnderTest, err = s3.NewBucket("does-not-exist", liveRegion, endpoint, creds, false)
					Expect(err).NotTo(HaveOccurred())
				})

				It("errors", func() {
					Expect(err).To(MatchError(ContainSubstring("failed to list files from bucket does-not-exist")))
				})

			})

			Context("when the bucket has a lot of files", func() {
				BeforeEach(func() {
					bucketObjectUnderTest, err = s3.NewBucket("sdk-unversioned-big-bucket-integration-test", liveRegion, endpoint, creds, false)
					Expect(err).NotTo(HaveOccurred())
				})

				It("works", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(len(files)).To(Equal(2001))
				})
			})

		})

		Context("When I ask for a list of files in a directory", func() {
			JustBeforeEach(func() {
				files, err = bucketObjectUnderTest.ListFiles("live/location")
			})

			It("should list all the files in the directory", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(files).To(ConsistOf("leaf/node"))
			})
		})
	})

	Describe("CopyObject", func() {
		var (
			backedUpFiles    []string
			backupBucketName string
		)

		BeforeEach(func() {
			backupBucketName = setUpUnversionedBucket(backupRegion, endpoint, creds)

			bucketObjectUnderTest, err = s3.NewBucket(backupBucketName, backupRegion, endpoint, creds, false)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			tearDownBucket(backupBucketName, endpoint, creds)
		})

		JustBeforeEach(func() {
			err = bucketObjectUnderTest.CopyObject(
				"leaf/node",
				"live/location",
				"2012_02_13_23_12_02/bucketIdFromDeployment",
				liveBucketName,
				liveRegion,
			)
		})

		It("copies the file to the backup bucket", func() {
			Expect(err).NotTo(HaveOccurred())
			backedUpFiles = listFiles(bucketObjectUnderTest.Name(), endpoint, creds)
			Expect(err).NotTo(HaveOccurred())
			Expect(backedUpFiles).To(ConsistOf([]string{"2012_02_13_23_12_02/bucketIdFromDeployment/leaf/node"}))
			Expect(getFileContents(
				backupBucketName,
				endpoint,
				"2012_02_13_23_12_02/bucketIdFromDeployment/leaf/node",
				creds),
			).To(Equal("CONTENTS"))
		})
	})
}
