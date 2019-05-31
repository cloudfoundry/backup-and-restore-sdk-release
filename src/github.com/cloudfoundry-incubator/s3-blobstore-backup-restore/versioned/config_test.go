package versioned_test

import (
	"github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/versioned"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Versioned", func() {
	Context("BuildVersionedBuckets", func() {
		It("builds versioned buckets from config", func() {
			configs := map[string]versioned.BucketConfig{
				"bucket1": {
					Name:               "live-name1",
					Region:             "live-region1",
					AwsAccessKeyId:     "my-id",
					AwsSecretAccessKey: "my-secret-key",
					Endpoint:           "my-s3-endpoint.aws",
					UseIAMProfile:      false,
				},
				"bucket2": {
					Name:               "live-name2",
					Region:             "live-region2",
					AwsAccessKeyId:     "my-id",
					AwsSecretAccessKey: "my-secret-key",
					Endpoint:           "my-s3-endpoint.aws",
					UseIAMProfile:      false,
				},
			}

			buckets, err := versioned.BuildVersionedBuckets(configs)

			Expect(err).NotTo(HaveOccurred())
			Expect(buckets).To(HaveLen(2))
			for _, n := range []string{"1", "2"} {
				Expect(buckets).To(HaveKey("bucket" + n))
				Expect(buckets["bucket"+n].Name()).To(Equal("live-name" + n))
				Expect(buckets["bucket"+n].Region()).To(Equal("live-region" + n))
			}
		})
	})
})
