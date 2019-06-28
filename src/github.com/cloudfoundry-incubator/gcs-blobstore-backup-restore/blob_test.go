package gcs_test

import (
	"github.com/cloudfoundry-incubator/backup-and-restore-sdk/gcs-blobstore-backup-restore"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Blob", func() {
	Context("with a normal blob", func() {
		Context("without a prefix", func() {
			It("has a name and suffix", func() {
				blob := gcs.NewBlob("some-name")
				Expect(blob.Name()).To(Equal("some-name"))
				Expect(blob.Resource()).To(Equal("some-name"))
			})
		})

		Context("with a prefix", func() {
			It("has a name and suffix", func() {
				blob := gcs.NewBlob("some-prefix/some-name")
				Expect(blob.Name()).To(Equal("some-prefix/some-name"))
				Expect(blob.Resource()).To(Equal("some-name"))
			})
		})
	})
})
