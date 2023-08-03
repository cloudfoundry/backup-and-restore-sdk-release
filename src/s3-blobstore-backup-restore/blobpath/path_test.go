package blobpath_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"s3-blobstore-backup-restore/blobpath"
)

var _ = Describe("Path", func() {
	It("joins", func() {
		Expect(blobpath.Join("prefix", "suffix")).To(Equal("prefix/suffix"))
	})

	It("trims a prefix", func() {
		Expect(blobpath.TrimPrefix("prefix/suffix", "prefix")).To(Equal("suffix"))
	})

	It("trims trailing delimiter", func() {
		Expect(blobpath.TrimTrailingDelimiter("path/")).To(Equal("path"))
	})
})
