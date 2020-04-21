package versioned_test

import (
	"s3-blobstore-backup-restore/s3bucket"
	"s3-blobstore-backup-restore/versioned"
	"s3-blobstore-backup-restore/versioned/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
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

		DescribeTable("passes the appropriate path/vhost information from the config to the bucket builder", func(forcePathStyle bool) {
			fakeBucket := new(fakes.FakeBucket)

			config := map[string]versioned.BucketConfig{
				"bucket": versioned.BucketConfig{ForcePathStyle: forcePathStyle},
			}

			forcePathStyles := []bool{}
			newBucketSpy := func(_, _, _ string, _ s3bucket.AccessKey, _, forcePathStyle bool) (versioned.Bucket, error) {
				forcePathStyles = append(forcePathStyles, forcePathStyle)
				return fakeBucket, nil
			}

			_, err := versioned.BuildVersionedBuckets(config, newBucketSpy)
			Expect(err).NotTo(HaveOccurred())
			Expect(forcePathStyles).To(HaveLen(1))
			Expect(forcePathStyles[0]).To(Equal(forcePathStyle), "forcePathStyle param to newBucket for bucket should match bucket config")
		},
			Entry("we force path style", true),
			Entry("we allow vhost style", false),
		)
	})
})
