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
		uploadFile(liveBucketName, awsEndpoint, "path1/file1", "FILE1", creds)
		uploadFile(liveBucketName, awsEndpoint, "live/location/leaf/node", "CONTENTS", creds)
		uploadFile(liveBucketName, awsEndpoint, "path2/file2", "FILE2", creds)
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
		BeforeEach(func() {
			var err error
			bucketObjectUnderTest, err = s3.NewBucket(liveBucketName, liveRegion, awsEndpoint, creds, false)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("When I ask for a list of all the files in the bucket", func() {
			It("should list all the files", func() {
				files, err := bucketObjectUnderTest.ListBlobs("")

				blob1 := s3.NewBlob("path1/file1")
				blob2 := s3.NewBlob("live/location/leaf/node")
				blob3 := s3.NewBlob("path2/file2")

				Expect(err).NotTo(HaveOccurred())
				Expect(files).To(ConsistOf(blob1, blob2, blob3))
			})

			Context("when s3 list-objects errors", func() {
				It("errors", func() {
					bucketObjectUnderTest, err := s3.NewBucket("does-not-exist", liveRegion, awsEndpoint, creds, false)
					Expect(err).NotTo(HaveOccurred())

					_, err = bucketObjectUnderTest.ListBlobs("")

					Expect(err).To(MatchError(ContainSubstring("failed to list files from bucket does-not-exist")))
				})
			})

			Context("when the bucket has a lot of files", func() {
				It("works", func() {
					bucketObjectUnderTest, err := s3.NewBucket("sdk-unversioned-big-bucket-integration-test", liveRegion, awsEndpoint, creds, false)
					Expect(err).NotTo(HaveOccurred())

					files, err := bucketObjectUnderTest.ListBlobs("")

					Expect(err).NotTo(HaveOccurred())
					Expect(len(files)).To(Equal(2001))
				})
			})
		})

		Context("When I ask for a list of files in a directory", func() {
			It("should list all the files in the directory", func() {
				files, err := bucketObjectUnderTest.ListBlobs("live/location")

				Expect(err).NotTo(HaveOccurred())
				Expect(files).To(ConsistOf(s3.NewBlob("leaf/node")))
			})
		})
	})
})
