package s3bucket_test

import (
	"s3-blobstore-backup-restore/s3bucket"

	"github.com/aws/aws-sdk-go/aws"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Creating a session on a bucket", func() {

	When("we want to use a path style bucket addresses", func() {
		It("adds the appropriate property to the config object", func() {

			session, err := s3bucket.CreateSessionImpl("", nil, "", true)

			Expect(err).NotTo(HaveOccurred())
			Expect(session.Config.S3ForcePathStyle).To(Equal(aws.Bool(true)))
		})
	})

	When("we want to use a v-host style bucket addresses", func(){
		It("adds the appropriate property to the config object", func() {

			session, err := s3bucket.CreateSessionImpl("", nil, "", false)

			Expect(err).NotTo(HaveOccurred())
			Expect(session.Config.S3ForcePathStyle).To(Equal(aws.Bool(false)))
		})
	})
})
