package s3bucket_test

import (
	"github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/s3bucket"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Bucket", func() {
	Describe("CheckIfVersioned", func() {
		Context("when the bucket is not versioned", func() {
			var unversionedBucketName string
			var unversionedBucketRegion string
			var endpoint string
			var creds s3bucket.AccessKey
			var err error
			var bucketObjectUnderTest s3bucket.VersionedBucket

			BeforeEach(func() {
				endpoint = ""
				unversionedBucketRegion = "eu-west-1"
				creds = s3bucket.AccessKey{
					Id:     TestAWSAccessKeyID,
					Secret: TestAWSSecretAccessKey,
				}

				unversionedBucketName = setUpUnversionedBucket(unversionedBucketRegion, endpoint, creds)
				uploadFile(unversionedBucketName, endpoint, "unversioned-test", "UNVERSIONED-TEST", creds)

				bucketObjectUnderTest, err = s3bucket.NewBucket(unversionedBucketName, unversionedBucketRegion, endpoint, creds, false)
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
