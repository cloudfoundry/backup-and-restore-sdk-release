package gcs_test

import (
	"github.com/cloudfoundry-incubator/gcs-blobstore-backup-restore"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Bucket", func() {
	Context("Buckets", func() {
		It("builds buckets", func() {
			config := map[string]gcs.Config{
				"droplets": {
					Name:              "droplets-bucket",
					ServiceAccountKey: "my-gcp-service-account-key",
				},
			}

			buckets := gcs.BuildBuckets(config)

			Expect(buckets).To(HaveLen(1))
			Expect(buckets["droplets"].Name()).To(Equal("droplets-bucket"))
		})
	})
})
