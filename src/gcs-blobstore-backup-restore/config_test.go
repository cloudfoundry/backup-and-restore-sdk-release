package gcs_test

import (
	"os"

	"gcs-blobstore-backup-restore"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
	Describe("ParseConfig", func() {
		Context("when the config file exists and is valid", func() {
			var configFile *os.File
			var err error
			var config map[string]gcs.Config

			BeforeEach(func() {
				configFile, err = os.CreateTemp("", "gcs_config")
				Expect(err).NotTo(HaveOccurred())
			})

			It("parses", func() {
				configJson := `{
				"bucket_id": {
					"bucket_name": "bucket",
					"backup_bucket_name": "backup_bucket"
				}
			}`
				os.WriteFile(configFile.Name(), []byte(configJson), 0644)

				config, err = gcs.ParseConfig(configFile.Name())

				Expect(config["bucket_id"]).To(Equal(gcs.Config{
					BucketName:       "bucket",
					BackupBucketName: "backup_bucket",
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
				configFile, err := os.CreateTemp("", "gcs_config")
				Expect(err).NotTo(HaveOccurred())
				os.WriteFile(configFile.Name(), []byte{}, 0000)

				_, err = gcs.ParseConfig(configFile.Name())

				Expect(err).To(MatchError(ContainSubstring("unexpected end of JSON input")))
			})
		})
	})

	Describe("ReadGCPServiceAccountKey", func() {
		Context("when the config file exists and is valid", func() {
			var configFile *os.File
			var err error
			var config string

			BeforeEach(func() {
				configFile, err = os.CreateTemp("", "gcs_config")
				Expect(err).NotTo(HaveOccurred())
			})

			It("reads", func() {
				configJson := `{"name":"value"}`
				os.WriteFile(configFile.Name(), []byte(configJson), 0644)

				config, err = gcs.ReadGCPServiceAccountKey(configFile.Name())

				Expect(config).To(Equal(`{"name":"value"}`))
			})
		})

		Context("when the config file does not exist", func() {
			It("returns an error", func() {
				_, err := gcs.ReadGCPServiceAccountKey("")

				Expect(err).To(HaveOccurred())
			})
		})
	})
})
