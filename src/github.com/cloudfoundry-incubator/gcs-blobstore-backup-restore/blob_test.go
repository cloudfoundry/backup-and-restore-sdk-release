package gcs_test

import (
	"github.com/cloudfoundry-incubator/gcs-blobstore-backup-restore"
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

			It("is a backup complete blob", func() {
				blob := gcs.NewBlob("some-name")
				Expect(blob.IsBackupComplete()).To(BeFalse())
			})
		})

		Context("with a prefix", func() {
			It("has a name and suffix", func() {
				blob := gcs.NewBlob("some-prefix/some-name")
				Expect(blob.Name()).To(Equal("some-prefix/some-name"))
				Expect(blob.Resource()).To(Equal("some-name"))
			})

			It("is a backup complete blob", func() {
				blob := gcs.NewBlob("some-prefix/some-name")
				Expect(blob.IsBackupComplete()).To(BeFalse())
			})
		})
	})

	Context("with a backup complete blob", func() {
		Context("without a prefix", func() {
			var blob gcs.Blob

			BeforeEach(func() {
				blob = gcs.NewBackupCompleteBlob("")
			})

			It("has a name and suffix", func() {
				Expect(blob.Name()).To(Equal("backup_complete"))
				Expect(blob.Resource()).To(Equal("backup_complete"))
			})

			It("is a backup complete blob", func() {
				Expect(blob.IsBackupComplete()).To(BeTrue())
			})
		})

		Context("with a prefix", func() {
			var blob gcs.Blob

			BeforeEach(func() {
				blob = gcs.NewBackupCompleteBlob("some-prefix")
			})

			It("has a name and suffix", func() {
				Expect(blob.Name()).To(Equal("some-prefix/backup_complete"))
				Expect(blob.Resource()).To(Equal("backup_complete"))
			})

			It("is a backup complete blob", func() {
				Expect(blob.IsBackupComplete()).To(BeTrue())
			})
		})
	})
})
