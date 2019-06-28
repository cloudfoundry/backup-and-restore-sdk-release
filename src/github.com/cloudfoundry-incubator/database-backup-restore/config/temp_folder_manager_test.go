package config_test

import (
	. "github.com/cloudfoundry-incubator/backup-and-restore-sdk/database-backup-restore/config"

	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TempFolderManager", func() {
	var tempFolderManager TempFolderManager
	var err error

	BeforeEach(func() {
		tempFolderManager, err = NewTempFolderManager()
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		tempFolderManager.Cleanup()
	})

	Describe("WriteTempFile", func() {
		It("creates a temp file with specified contents", func() {
			filePath, err := tempFolderManager.WriteTempFile("test contents")
			Expect(err).NotTo(HaveOccurred())
			Expect(filePath).To(BeAnExistingFile())
			Expect(ioutil.ReadFile(filePath)).To(Equal([]byte("test contents")))
		})
	})

	Describe("Cleanup", func() {
		It("removes any temp files created so far", func() {
			filePath, err := tempFolderManager.WriteTempFile("test contents")
			Expect(err).NotTo(HaveOccurred())
			Expect(filePath).To(BeAnExistingFile())
			Expect(ioutil.ReadFile(filePath)).To(Equal([]byte("test contents")))

			tempFolderManager.Cleanup()
			Expect(filePath).NotTo(BeAnExistingFile())
		})
	})
})
