package s3bucket_test

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"s3-blobstore-backup-restore/s3bucket"
)

var _ = Describe("Creating an S3 Client", func() {
	var creds aws.CredentialsProvider
	var options *s3.Options
	var spy = func(o *s3.Options) {
		options = o
		creds = o.Credentials
	}

	BeforeEach(func() {
		options = nil
		creds = nil
	})

	When("we are not using an IAMProfile", func() {
		It("provides static credentials", func() {
			_, err := s3bucket.NewBucket("", "", "", s3bucket.AccessKey{Id: "user", Secret: "pass"}, false, false, spy)

			Expect(err).NotTo(HaveOccurred())
			Expect(creds).To(BeAssignableToTypeOf(&aws.CredentialsCache{}))
			Expect(creds.(*aws.CredentialsCache).IsCredentialsProvider(&credentials.StaticCredentialsProvider{})).To(BeTrue())
		})
	})

	When("we are using an IAMProfile", func() {
		It("provides EC2 Role credentials", func() {
			_, err := s3bucket.NewBucket("", "", "", s3bucket.AccessKey{Id: "user", Secret: "pass"}, true, false, spy)

			Expect(err).NotTo(HaveOccurred())
			Expect(creds).To(BeAssignableToTypeOf(&aws.CredentialsCache{}))
			Expect(creds.(*aws.CredentialsCache).IsCredentialsProvider(&ec2rolecreds.Provider{})).To(BeTrue())
		})
	})

	When("we are using an assumed role", func() {
		It("provides the assumed role credentials", func() {
			_, err := s3bucket.NewBucketWithRoleARN("", "", "", "someRole", s3bucket.AccessKey{Id: "C0FFEEC0C0ABEAD", Secret: "DEAFD011"}, false, false, spy)

			Expect(err).NotTo(HaveOccurred())
			Expect(creds).To(BeAssignableToTypeOf(&stscreds.AssumeRoleProvider{}))
		})
	})

	Context("bucket addresses", func() {

		When("we want to use a path style", func() {
			It("adds the appropriate property to the config object", func() {
				_, err := s3bucket.NewBucket("", "", "", s3bucket.AccessKey{}, false, true, spy)

				Expect(err).NotTo(HaveOccurred())
				Expect(options).NotTo(BeNil())
				Expect(options.UsePathStyle).To(BeTrue())
			})
		})

		When("we want to use a v-host style", func() {
			It("adds the appropriate property to the config object", func() {

				_, err := s3bucket.NewBucket("", "", "", s3bucket.AccessKey{}, false, false, spy)

				Expect(err).NotTo(HaveOccurred())
				Expect(options).NotTo(BeNil())
				Expect(options.UsePathStyle).To(BeFalse())
			})
		})
	})

	Describe("Determining blob size", func() {
		When("the config specifies path style", func() {
			It("Uses path style property for the client", func() {
				bucket, err := s3bucket.NewBucket("fred", "", "", s3bucket.AccessKey{}, false, true, spy)
				Expect(err).NotTo(HaveOccurred())

				_, _ = bucket.GetBlobSizeImpl("", "", "", "")

				Expect(options.UsePathStyle).To(BeTrue())
			})
		})

		When("the config specifies vhost style", func() {
			It("uses vhost property for the client", func() {
				bucket, err := s3bucket.NewBucket("fred", "", "", s3bucket.AccessKey{}, false, false, spy)
				Expect(err).NotTo(HaveOccurred())

				_, _ = bucket.GetBlobSizeImpl("", "", "", "")

				Expect(options.UsePathStyle).To(BeFalse())
			})
		})
	})
})
