package s3_test

import (
	"github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/s3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("IncrementalBucket", func() {
	const (
		liveRegion  = "eu-west-1"
		awsEndpoint = ""
	)

	var (
		liveBucketName        string
		bucketObjectUnderTest s3.Bucket
		creds                 s3.AccessKey
	)

	BeforeEach(func() {
		creds = s3.AccessKey{Id: TestAWSAccessKeyID, Secret: TestAWSSecretAccessKey}

		liveBucketName = setUpUnversionedBucket(liveRegion, awsEndpoint, creds)
		uploadFile(liveBucketName, awsEndpoint, "path1/blob1", "blob1-content", creds)
		uploadFile(liveBucketName, awsEndpoint, "live/location/leaf/node", "", creds)
		uploadFile(liveBucketName, awsEndpoint, "path2/blob2", "", creds)

		var err error
		bucketObjectUnderTest, err = s3.NewBucket(liveBucketName, liveRegion, awsEndpoint, creds, false)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		tearDownBucket(liveBucketName, awsEndpoint, creds)
	})

	//It("implements the incremental bucket interface", func() {
	//	bucket, err := s3.NewBucket("", "", "", s3.AccessKey{}, false)
	//	Expect(err).NotTo(HaveOccurred())
	//
	//	_ = incremental.BucketPair{BackupBucket: bucket}
	//})

	Describe("ListBlobs", func() {
		Context("without a prefix", func() {
			It("lists all the blobs in the bucket", func() {
				blobs, err := bucketObjectUnderTest.ListBlobs("")

				blob1 := s3.NewBlob("path1/blob1")
				blob2 := s3.NewBlob("live/location/leaf/node")
				blob3 := s3.NewBlob("path2/blob2")

				Expect(err).NotTo(HaveOccurred())
				Expect(blobs).To(ConsistOf(blob1, blob2, blob3))
			})

			Context("when s3 list-objects errors", func() {
				It("errors", func() {
					bucketObjectUnderTest, err := s3.NewBucket("does-not-exist", liveRegion, awsEndpoint, creds, false)
					Expect(err).NotTo(HaveOccurred())

					_, err = bucketObjectUnderTest.ListBlobs("")

					Expect(err).To(MatchError(ContainSubstring("failed to list blobs from bucket does-not-exist")))
				})
			})

			Context("when the bucket has a lot of blobs", func() {
				It("works", func() {
					bucketObjectUnderTest, err := s3.NewBucket("sdk-unversioned-big-bucket-integration-test", liveRegion, awsEndpoint, creds, false)
					Expect(err).NotTo(HaveOccurred())

					blobs, err := bucketObjectUnderTest.ListBlobs("")

					Expect(err).NotTo(HaveOccurred())
					Expect(len(blobs)).To(Equal(2001))
				})
			})
		})

		Context("with a prefix", func() {
			It("lists all the blobs in the directory", func() {
				blobs, err := bucketObjectUnderTest.ListBlobs("live/location")

				Expect(err).NotTo(HaveOccurred())
				Expect(blobs).To(ConsistOf(s3.NewBlob("leaf/node")))
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
				dirs, err := bucketObjectUnderTest.ListDirectories()

				Expect(err).NotTo(HaveOccurred())
				Expect(dirs).To(ConsistOf("path1", "live", "path2"))
			})
		})

		Context("when s3 list-objects errors", func() {
			It("errors", func() {
				bucketObjectUnderTest, err := s3.NewBucket("does-not-exist", liveRegion, awsEndpoint, creds, false)
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

			err := bucketObjectUnderTest.CopyBlobWithinBucket("path1/blob1", copyPath)

			Expect(err).NotTo(HaveOccurred())
			actualContents := getFileContents(liveBucketName, awsEndpoint, copyPath, creds)
			Expect(actualContents).To(Equal("blob1-content"))
		})
	})
})
