package azure_test

import (
	"io/ioutil"

	"github.com/cloudfoundry-incubator/azure-blobstore-backup-restore"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ParseConfig", func() {
	Context("when the config file exists and is valid", func() {
		It("parses it", func() {
			configFile, err := ioutil.TempFile("", "azure_config")
			Expect(err).NotTo(HaveOccurred())

			configJson := `{
				"container_id": {
					"name": "container_name",
					"azure_account_name": "ACCOUNT_NAME",
					"azure_account_key": "ACCOUNT_KEY"
				}
			}`
			ioutil.WriteFile(configFile.Name(), []byte(configJson), 0644)

			config, err := azure.ParseConfig(configFile.Name())

			Expect(config["container_id"]).To(Equal(azure.ContainerConfig{
				Name:             "container_name",
				AzureAccountName: "ACCOUNT_NAME",
				AzureAccountKey:  "ACCOUNT_KEY",
			}))
		})
	})

	Context("when the config file does not exist", func() {
		It("returns an error", func() {
			_, err := azure.ParseConfig("")

			Expect(err).To(HaveOccurred())
		})
	})

	Context("when the config file is not valid", func() {
		It("returns an error", func() {
			configFile, err := ioutil.TempFile("", "azure_config")
			Expect(err).NotTo(HaveOccurred())
			ioutil.WriteFile(configFile.Name(), []byte{}, 0000)

			_, err = azure.ParseConfig(configFile.Name())

			Expect(err).To(MatchError(ContainSubstring("unexpected end of JSON input")))
		})
	})
})
