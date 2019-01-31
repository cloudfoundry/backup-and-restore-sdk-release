package s3_test

import (
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
			TestAWSAccessKeyID,
			TestAWSSecretAccessKey,
		)
	})

	Describe("ECS S3-compatible buckets", func() {
		RunUnversionedBucketTests(
			"eu-west-1",
			"us-west-1",
			"https://object.ecstestdrive.com",
			TestECSAccessKeyID,
			TestECSSecretAccessKey,
		)
	})

	Describe("AWS S3 large file test", func() {
		RunLargeFileTest(
			"eu-west-1", "",
			TestAWSAccessKeyID,
			TestAWSSecretAccessKey,
			"large-blob-test-bucket-unversioned")
	})

	Describe("ECS S3-compatible buckets large file test", func() {
		RunLargeFileTest(
			"eu-west-1",
			"https://object.ecstestdrive.com",
			TestECSAccessKeyID,
			TestECSSecretAccessKey,
			"sdk-unversioned-big-file-unit-test",
		)
	})
})

func RunLargeFileTest(region, endpoint, accessKey, secretKey, preExistingBigFileBucketName string) {

	var (
		creds                 s3.AccessKey
		bucketObjectUnderTest s3.UnversionedBucket
		err                   error
		destinationBucketName string
	)

	BeforeEach(func() {
		creds = s3.AccessKey{
			Id:     accessKey,
			Secret: secretKey,
		}

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
			Equal("91d50642dd930e9542c39d36f0516d45f4e1af0d"))
	})
}

func RunUnversionedBucketTests(liveRegion, backupRegion, endpoint, accessKey, secretKey string) {
	var (
		liveBucketName        string
		bucketObjectUnderTest s3.UnversionedBucket
		err                   error
		creds                 s3.AccessKey
	)

	BeforeEach(func() {
		creds = s3.AccessKey{Id: accessKey, Secret: secretKey}

		liveBucketName = setUpUnversionedBucket(liveRegion, endpoint, creds)
		uploadFile(liveBucketName, endpoint, "path1/file1", "FILE1", creds)
		uploadFile(liveBucketName, endpoint, "live/location/leaf/node", "CONTENTS", creds)
		uploadFile(liveBucketName, endpoint, "path2/file2", "FILE2", creds)
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
					Expect(err).To(MatchError(ContainSubstring("failed to list blobs from bucket does-not-exist")))
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
