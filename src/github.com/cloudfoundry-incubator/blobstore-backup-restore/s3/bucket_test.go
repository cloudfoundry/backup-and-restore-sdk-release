package s3_test

import (
	"os"

	"github.com/cloudfoundry-incubator/blobstore-backup-restore/s3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Bucket", func() {

	Describe("IsVersioned", func() {
		Context("when the bucket is not versioned", func() {

			var unversionedBucketName string
			var unversionedBucketRegion string
			var endpoint string
			var creds s3.S3AccessKey
			var err error
			var bucketObjectUnderTest s3.VersionedBucket

			BeforeEach(func() {
				endpoint = ""
				unversionedBucketRegion = "eu-west-1"
				creds = s3.S3AccessKey{
					Id:     os.Getenv("TEST_AWS_ACCESS_KEY_ID"),
					Secret: os.Getenv("TEST_AWS_SECRET_ACCESS_KEY"),
				}

				unversionedBucketName = setUpUnversionedBucket(unversionedBucketRegion, endpoint, creds)
				uploadFile(unversionedBucketName, endpoint, "unversioned-test", "UNVERSIONED-TEST", creds)

				bucketObjectUnderTest, err = s3.NewBucket(unversionedBucketName, unversionedBucketRegion, endpoint, creds)
				Expect(err).NotTo(HaveOccurred())
			})

			JustBeforeEach(func() {
				err = bucketObjectUnderTest.CheckIfVersioned()
			})

			AfterEach(func() {
				tearDownBucket(unversionedBucketName, endpoint, creds)
			})

			It("fails", func() {
				Expect(err).To(MatchError(ContainSubstring("is not versioned")))
			})
		})
	})

})
