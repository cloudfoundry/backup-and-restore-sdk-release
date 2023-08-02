package gcs_test

import (
	"gcs-blobstore-backup-restore"

	. "github.com/onsi/ginkgo/v2"
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
