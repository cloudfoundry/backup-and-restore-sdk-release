package azure_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/azure-blobstore-backup-restore"
)

var _ = Describe("ContainerBuilder", func() {
	Context("Containers", func() {
		It("builds live containers", func() {
			config := map[string]azure.ContainerConfig{
				"droplets": {
					Name:           "droplets-container",
					StorageAccount: "droplets-account",
					StorageKey:     "ZHJvcGxldHMta2V5",
					Environment:    azure.DefaultEnvironment,
				},
			}
			builder := azure.NewContainerBuilder(config)

			containers, err := builder.Containers()

			Expect(err).NotTo(HaveOccurred())
			Expect(containers).To(HaveLen(1))
			Expect(containers["droplets"].Name()).To(Equal("droplets-container"))
		})
	})

	Context("RestoreFromStorageAccounts", func() {
		It("builds restore from containers", func() {
			builder := azure.NewContainerBuilder(map[string]azure.ContainerConfig{
				"droplets": {
					Environment: azure.DefaultEnvironment,
					RestoreFrom: azure.RestoreFromConfig{
						StorageAccount: "restore-from-account",
						StorageKey:     "restore-from-key",
					},
				},
				"buildpacks": {
					Name:           "buildpacks-container",
					StorageAccount: "buildpacks-account",
					StorageKey:     "buildpacks-key",
					Environment:    azure.DefaultEnvironment,
				},
			})

			storageAccounts := builder.RestoreFromStorageAccounts()

			Expect(storageAccounts).To(HaveLen(1))
			Expect(storageAccounts["droplets"].Name).To(Equal("restore-from-account"))
			Expect(storageAccounts["droplets"].Key).To(Equal("restore-from-key"))
		})
	})
})
