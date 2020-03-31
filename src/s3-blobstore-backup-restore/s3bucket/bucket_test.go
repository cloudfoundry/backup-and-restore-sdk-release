package s3bucket_test

import (
	"s3-blobstore-backup-restore/s3bucket"

	"github.com/aws/aws-sdk-go/aws"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Bucket", func() {

	Context("Given a Path Style", func() {
		FIt("adds that property to the config object", func() {
			config := &aws.Config{
				Region:           aws.String("foo"),
				Credentials:      nil,
				Endpoint:         aws.String("endpoint"),
				S3ForcePathStyle: aws.Bool(true),
			}

			session, _ := s3bucket.CreateSession(
				*config.Region,
				config.Credentials,
				*config.Endpoint,
				*config.S3ForcePathStyle)

			Expect(session.Config.S3ForcePathStyle).To(Equal(aws.Bool(false)))
		})
	})

})
