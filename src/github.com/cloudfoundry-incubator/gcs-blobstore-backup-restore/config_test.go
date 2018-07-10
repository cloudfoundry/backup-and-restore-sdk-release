package gcs_test

import (
	"io/ioutil"

	"os"

	"github.com/cloudfoundry-incubator/gcs-blobstore-backup-restore"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ParseConfig", func() {
	Context("when the config file exists and is valid", func() {
		var configFile *os.File
		var err error
		var config map[string]gcs.Config

		BeforeEach(func() {
			configFile, err = ioutil.TempFile("", "gcs_config")
			Expect(err).NotTo(HaveOccurred())
		})

		It("parses", func() {
			configJson := `{
				"bucket_id": {
					"name": "bucket_name",
					"gcp_service_account_key": "my-service-account-key"
				}
			}`
			ioutil.WriteFile(configFile.Name(), []byte(configJson), 0644)

			config, err = gcs.ParseConfig(configFile.Name())

			Expect(config["bucket_id"]).To(Equal(gcs.Config{
				Name:              "bucket_name",
				ServiceAccountKey: "my-service-account-key",
			}))
		})
	})

	Context("when the config file does not exist", func() {
		It("returns an error", func() {
			_, err := gcs.ParseConfig("")

			Expect(err).To(HaveOccurred())
		})
	})

	Context("when the config file is not valid", func() {
		It("returns an error", func() {
			configFile, err := ioutil.TempFile("", "gcs_config")
			Expect(err).NotTo(HaveOccurred())
			ioutil.WriteFile(configFile.Name(), []byte{}, 0000)

			_, err = gcs.ParseConfig(configFile.Name())

			Expect(err).To(MatchError(ContainSubstring("unexpected end of JSON input")))
		})
	})
})
