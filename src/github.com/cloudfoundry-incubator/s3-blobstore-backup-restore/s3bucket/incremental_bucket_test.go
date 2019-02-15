package s3bucket_test

import (
	"fmt"

	"github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/s3bucket"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("IncrementalBucket", func() {
	Describe("AWS S3 buckets", func() {
		RunIncrementalBucketContractTests(
			"eu-west-1",
			"eu-central-1",
			"",
			TestAWSAccessKeyID,
			TestAWSSecretAccessKey,
		)
	})

	XDescribe("ECS S3-compatible buckets", func() {
		RunIncrementalBucketContractTests(
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

	XDescribe("ECS S3-compatible buckets large file test", func() {
		RunLargeFileTest(
			"eu-west-1",
			"https://object.ecstestdrive.com",
			TestECSAccessKeyID,
			TestECSSecretAccessKey,
			"sdk-unversioned-big-file-unit-test",
		)
	})
})

func RunIncrementalBucketContractTests(liveRegion, backupRegion, awsEndpoint, accessKey, secretKey string) {
	var (
		liveBucketName string
		liveBucket     s3bucket.Bucket
		creds          s3bucket.AccessKey
	)

	BeforeEach(func() {
		creds = s3bucket.AccessKey{Id: accessKey, Secret: secretKey}

		liveBucketName = setUpUnversionedBucket(liveRegion, awsEndpoint, creds)
		uploadFile(liveBucketName, awsEndpoint, "path1/blob1", "blob1-content", creds)
		uploadFile(liveBucketName, awsEndpoint, "live/location/leaf/node", "", creds)
		uploadFile(liveBucketName, awsEndpoint, "path2/blob2", "", creds)

		var err error
		liveBucket, err = s3bucket.NewBucket(liveBucketName, liveRegion, awsEndpoint, creds, false)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		tearDownBucket(liveBucketName, awsEndpoint, creds)
	})

	Describe("ListBlobs", func() {
		Context("without a prefix", func() {
			It("lists all the blobs in the bucket", func() {
				blobs, err := liveBucket.ListBlobs("")

				blob1 := s3bucket.NewBlob("path1/blob1")
				blob2 := s3bucket.NewBlob("live/location/leaf/node")
				blob3 := s3bucket.NewBlob("path2/blob2")

				Expect(err).NotTo(HaveOccurred())
				Expect(blobs).To(ConsistOf(blob1, blob2, blob3))
			})

			Context("when s3 list-objects errors", func() {
				It("errors", func() {
					bucketObjectUnderTest, err := s3bucket.NewBucket("does-not-exist", liveRegion, awsEndpoint, creds, false)
					Expect(err).NotTo(HaveOccurred())

					_, err = bucketObjectUnderTest.ListBlobs("")

					Expect(err).To(MatchError(ContainSubstring("failed to list blobs from bucket does-not-exist")))
				})
			})

			Context("when the bucket has a lot of blobs", func() {
				It("works", func() {
					bucketObjectUnderTest, err := s3bucket.NewBucket("sdk-unversioned-big-bucket-integration-test", liveRegion, awsEndpoint, creds, false)
					Expect(err).NotTo(HaveOccurred())

					blobs, err := bucketObjectUnderTest.ListBlobs("")

					Expect(err).NotTo(HaveOccurred())
					Expect(len(blobs)).To(Equal(2001))
				})
			})
		})

		Context("with a prefix", func() {
			It("lists all the blobs in the directory", func() {
				blobs, err := liveBucket.ListBlobs("live/location")

				Expect(err).NotTo(HaveOccurred())
				Expect(blobs).To(ConsistOf(s3bucket.NewBlob("leaf/node")))
			})
		})
	})

	Describe("ListDirectories", func() {
		Context("when there are several directories", func() {
			BeforeEach(func() {
				uploadFile(liveBucketName, awsEndpoint, "path1/another-blob", "", creds)
				uploadFile(liveBucketName, awsEndpoint, "top-level-blob", "", creds)
			})

			It("lists the top-level directories", func() {
				dirs, err := liveBucket.ListDirectories()

				Expect(err).NotTo(HaveOccurred())
				Expect(dirs).To(ConsistOf("path1", "live", "path2"))
			})
		})

		Context("when s3 list-objects errors", func() {
			It("errors", func() {
				bucketObjectUnderTest, err := s3bucket.NewBucket("does-not-exist", liveRegion, awsEndpoint, creds, false)
				Expect(err).NotTo(HaveOccurred())

				_, err = bucketObjectUnderTest.ListDirectories()
				Expect(err).To(MatchError(ContainSubstring("failed to list directories from bucket does-not-exist")))
			})
		})
	})

	Describe("CopyBlobWithinBucket", func() {
		It("copies the blob", func() {
			blobs := listFiles(liveBucketName, awsEndpoint, creds)
			copyPath := "path1/blob1-copy"
			Expect(blobs).NotTo(ContainElement(copyPath))

			err := liveBucket.CopyBlobWithinBucket("path1/blob1", copyPath)

			Expect(err).NotTo(HaveOccurred())
			actualContents := getFileContents(liveBucketName, awsEndpoint, copyPath, creds)
			Expect(actualContents).To(Equal("blob1-content"))
		})

		Context("when the blob does not exist", func() {
			It("errors", func() {
				err := liveBucket.CopyBlobWithinBucket("does-not-exist", "copy-of-does-not-exist")

				Expect(err).To(MatchError(
					ContainSubstring(fmt.Sprintf("failed to get blob size for blob 'does-not-exist' in bucket '%s'", liveBucketName)),
				))
			})
		})
	})

	Describe("CopyBlobFromBucket", func() {
		var (
			backupBucketName string
			backupBucket     s3bucket.Bucket
		)

		BeforeEach(func() {
			creds = s3bucket.AccessKey{Id: accessKey, Secret: secretKey}
			backupBucketName = setUpUnversionedBucket(backupRegion, awsEndpoint, creds)
			var err error
			backupBucket, err = s3bucket.NewBucket(backupBucketName, backupRegion, awsEndpoint, creds, false)

			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			tearDownBucket(backupBucketName, awsEndpoint, creds)
		})

		It("copies the blob", func() {
			blobPath := "path1/blob1"

			err := backupBucket.CopyBlobFromBucket(liveBucket, blobPath, blobPath)

			Expect(err).NotTo(HaveOccurred())
			actualContents := getFileContents(backupBucketName, awsEndpoint, blobPath, creds)
			Expect(actualContents).To(Equal("blob1-content"))
		})

		Context("when the blob does not exist", func() {
			It("errors", func() {
				err := backupBucket.CopyBlobFromBucket(liveBucket, "does-not-exist", "copy-of-does-not-exist")

				Expect(err).To(MatchError(
					ContainSubstring(fmt.Sprintf("failed to get blob size for blob 'does-not-exist' in bucket '%s'", liveBucketName)),
				))
			})
		})
	})

	Describe("UploadBlob", func() {
		It("uploads the blob", func() {
			blobPath := "some/blob"

			err := liveBucket.UploadBlob(blobPath, "blob contents")

			Expect(err).NotTo(HaveOccurred())
			actualContents := getFileContents(liveBucketName, awsEndpoint, blobPath, creds)
			Expect(actualContents).To(Equal("blob contents"))
		})

		Context("when the bucket does not exist", func() {
			It("errors", func() {
				bucket, err := s3bucket.NewBucket("does-not-exist", liveRegion, awsEndpoint, creds, false)
				Expect(err).NotTo(HaveOccurred())

				err = bucket.UploadBlob("some/blob", "blob contents")

				Expect(err).To(MatchError(ContainSubstring("failed to upload blob")))
			})
		})
	})

	Describe("HasBlob", func() {
		Context("when the blob exists", func() {
			It("returns true", func() {
				exists, err := liveBucket.HasBlob("path1/blob1")
				Expect(err).NotTo(HaveOccurred())

				Expect(exists).To(BeTrue())
			})
		})

		Context("when the blob does not exist", func() {
			It("returns false", func() {
				exists, err := liveBucket.HasBlob("does-not-exist-blob")
				Expect(err).NotTo(HaveOccurred())

				Expect(exists).To(BeFalse())
			})
		})

		Context("when the bucket does not exist", func() {
			It("errors", func() {
				bucket, err := s3bucket.NewBucket("does-not-exist", liveRegion, awsEndpoint, creds, false)
				Expect(err).NotTo(HaveOccurred())
				_, err = bucket.HasBlob("does-not-exist-blob")

				Expect(err).To(MatchError(ContainSubstring("failed to check if blob exists")))
			})
		})
	})
}

func RunLargeFileTest(region, endpoint, accessKey, secretKey, preExistingBigFileBucketName string) {
	var (
		creds      s3bucket.AccessKey
		bucketName string
	)

	BeforeEach(func() {
		creds = s3bucket.AccessKey{
			Id:     accessKey,
			Secret: secretKey,
		}

		bucketName = setUpUnversionedBucket("eu-west-1", endpoint, creds)
	})

	AfterEach(func() {
		tearDownBucket(bucketName, endpoint, creds)
	})

	It("works", func() {
		bucket, err := s3bucket.NewBucket(bucketName, region, endpoint, creds, false)
		Expect(err).NotTo(HaveOccurred())

		srcBucket, err := s3bucket.NewBucket(preExistingBigFileBucketName, region, endpoint, creds, false)
		Expect(err).NotTo(HaveOccurred())

		err = bucket.CopyBlobFromBucket(srcBucket, "big_file", "path/to/big_file")

		By("succeeding")
		Expect(err).NotTo(HaveOccurred())

		By("copying the large file")
		Expect(listFiles(bucketName, endpoint, creds)).To(ConsistOf("path/to/big_file"))

		By("not corrupting the large file")
		Expect(
			shasum(downloadFileToTmp(bucketName, endpoint, "path/to/big_file", creds))).To(
			Equal("91d50642dd930e9542c39d36f0516d45f4e1af0d"))
	})
}
