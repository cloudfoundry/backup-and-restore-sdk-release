package versioned_test

import (

	"s3-blobstore-backup-restore/versioned"
	"s3-blobstore-backup-restore/s3bucket"
	"s3-blobstore-backup-restore/versioned/fakes"
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

			buckets, err := versioned.BuildVersionedBuckets(configs, versioned.NewVersionedBucket)

			Expect(err).NotTo(HaveOccurred())
			Expect(buckets).To(HaveLen(2))
			for _, n := range []string{"1", "2"} {
				Expect(buckets).To(HaveKey("bucket" + n))
				Expect(buckets["bucket"+n].Name()).To(Equal("live-name" + n))
				Expect(buckets["bucket"+n].Region()).To(Equal("live-region" + n))
			}
		})

		It("passes the appropriate path/vhost information from the config to the bucket builder", func() {
			fakeBucket := new(fakes.FakeBucket)

			configs := map[string]versioned.BucketConfig{
				"bucket1": versioned.BucketConfig{ ForcePathStyle: true },
				"bucket2": versioned.BucketConfig{ ForcePathStyle: false},
			}

			forcePathStyles := []bool{}
			newBucketSpy := func(_, _, _ string, _ s3bucket.AccessKey, _, forcePathStyle bool) (versioned.Bucket, error) {
				forcePathStyles = append(forcePathStyles, forcePathStyle)
				return fakeBucket, nil
			}

			_, err := versioned.BuildVersionedBuckets(configs, newBucketSpy)
			Expect(err).NotTo(HaveOccurred())

			Expect(forcePathStyles[0]).To(BeTrue(), "forcePathStyle param to newBucket for bucket1 should match bucket config")
			Expect(forcePathStyles[1]).To(BeFalse(), "forcePathStyle param to newBucket for bucket2 should match bucket config")
		})
	})
})
