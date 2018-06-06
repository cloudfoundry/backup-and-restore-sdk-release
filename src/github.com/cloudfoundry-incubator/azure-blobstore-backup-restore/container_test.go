package azure_test

import (
	"strconv"
	"time"

	"fmt"

	"github.com/cloudfoundry-incubator/azure-blobstore-backup-restore"
	. "github.com/cloudfoundry-incubator/azure-blobstore-backup-restore/system_tests/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Container", func() {
	var azureClient AzureClient
	var container azure.Container
	var err error
	var eTag1, eTag2, eTag3 string
	var fileName1, fileName2, fileName3 string

	BeforeEach(func() {
		azureClient = NewAzureClient(MustHaveEnv("AZURE_STORAGE_ACCOUNT"), MustHaveEnv("AZURE_STORAGE_KEY"))
	})

	Describe("NewSDKContainer", func() {
		Context("when the account name is invalid", func() {
			It("returns an error", func() {
				container, err = azure.NewSDKContainer("", "\n", "", azure.DefaultEnvironment)

				Expect(err).To(MatchError("invalid account name: '\n'"))
				Expect(container).To(Equal(azure.SDKContainer{}))
			})
		})

		Context("when the account key is not valid base64", func() {
			It("returns an error", func() {
				container, err := azure.NewSDKContainer("", "", "#", azure.DefaultEnvironment)

				Expect(err).To(MatchError(ContainSubstring("invalid storage key: '")))
				Expect(container).To(Equal(azure.SDKContainer{}))
			})
		})

		Context("when the environment is not valid", func() {
			It("returns an error", func() {
				container, err := azure.NewSDKContainer("", "", "", "not-valid-environment")

				Expect(err).To(MatchError(ContainSubstring("invalid environment: not-valid-environment")))
				Expect(container).To(Equal(azure.SDKContainer{}))
			})
		})
	})

	Describe("Name", func() {
		It("returns the container name", func() {
			name := "container-name"

			container, err = azure.NewSDKContainer(name, "", "", azure.DefaultEnvironment)

			Expect(err).NotTo(HaveOccurred())
			Expect(container.Name()).To(Equal(name))
		})
	})

	Describe("SoftDeleteEnabled", func() {
		Context("when soft delete is enabled on the container's storage service", func() {
			BeforeEach(func() {
				containerName := azureClient.CreateContainerWithUniqueName("sdk-azure-test-")
				container = newContainer(containerName)
			})

			AfterEach(func() {
				azureClient.DeleteContainer(container.Name())
			})

			It("returns true", func() {
				enabled, err := container.SoftDeleteEnabled()

				Expect(err).NotTo(HaveOccurred())
				Expect(enabled).To(BeTrue())
			})
		})

		Context("when soft delete is disabled on the container's storage service", func() {
			It("returns false", func() {
				container, err = azure.NewSDKContainer(
					"",
					MustHaveEnv("AZURE_STORAGE_ACCOUNT_NO_SOFT_DELETE"),
					MustHaveEnv("AZURE_STORAGE_KEY_NO_SOFT_DELETE"),
					azure.DefaultEnvironment,
				)

				enabled, err := container.SoftDeleteEnabled()

				Expect(err).NotTo(HaveOccurred())
				Expect(enabled).To(BeFalse())
			})
		})

		Context("when retrieving the storage service properties fails", func() {
			It("returns an error", func() {
				container, err = azure.NewSDKContainer("", "", "", azure.DefaultEnvironment)

				_, err = container.SoftDeleteEnabled()

				Expect(err.Error()).To(ContainSubstring("failed fetching properties for storage account: '"))
			})
		})
	})

	Describe("ListBlobs", func() {
		Context("when the container has a few files and snapshots", func() {
			BeforeEach(func() {
				containerName := azureClient.CreateContainerWithUniqueName("sdk-azure-test-")
				container = newContainer(containerName)

				fileName1 = "test_file_1_" + strconv.FormatInt(time.Now().Unix(), 10)
				fileName2 = "test_file_2_" + strconv.FormatInt(time.Now().Unix(), 10)
				fileName3 = "test_file_3_" + strconv.FormatInt(time.Now().Unix(), 10)

				azureClient.WriteFileInContainer(container.Name(), fileName1, "TEST_BLOB_1_OLD")
				eTag1 = azureClient.WriteFileInContainer(container.Name(), fileName1, "TEST_BLOB_1")
				azureClient.WriteFileInContainer(container.Name(), fileName2, "TEST_BLOB_2_OLDEST")
				azureClient.WriteFileInContainer(container.Name(), fileName2, "TEST_BLOB_2_OLD")
				eTag2 = azureClient.WriteFileInContainer(container.Name(), fileName2, "TEST_BLOB_2")
				eTag3 = azureClient.WriteFileInContainer(container.Name(), fileName3, "TEST_BLOB_3")
			})

			AfterEach(func() {
				azureClient.DeleteContainer(container.Name())
			})

			It("returns a list of containers with files and their etags", func() {
				blobs, err := container.ListBlobs()

				Expect(err).NotTo(HaveOccurred())
				Expect(blobs).To(Equal([]azure.BlobId{
					{Name: fileName1, ETag: eTag1},
					{Name: fileName2, ETag: eTag2},
					{Name: fileName3, ETag: eTag3},
				}))
			})
		})

		Context("when the container has a lots of files", func() {
			It("paginates correctly", func() {
				container, err := azure.NewSDKContainer(
					MustHaveEnv("AZURE_CONTAINER_NAME_MANY_FILES"),
					MustHaveEnv("AZURE_STORAGE_ACCOUNT"),
					MustHaveEnv("AZURE_STORAGE_KEY"),
					azure.DefaultEnvironment,
				)
				Expect(err).NotTo(HaveOccurred())

				blobs, err := container.ListBlobs()

				Expect(err).NotTo(HaveOccurred())
				Expect(len(blobs)).To(Equal(10104))
			})
		})

		Context("when listing the blobs fails", func() {
			It("returns an error", func() {
				container, err := azure.NewSDKContainer(
					"NON-EXISTENT-CONTAINER",
					MustHaveEnv("AZURE_STORAGE_ACCOUNT"),
					MustHaveEnv("AZURE_STORAGE_KEY"),
					azure.DefaultEnvironment,
				)

				_, err = container.ListBlobs()

				Expect(err.Error()).To(ContainSubstring("failed listing blobs in container 'NON-EXISTENT-CONTAINER':"))
			})
		})
	})

	Describe("CopyBlobsFrom", func() {
		BeforeEach(func() {
			containerName := azureClient.CreateContainerWithUniqueName("sdk-azure-test-")
			container = newContainer(containerName)
			fileName1 = "test_file_1_" + strconv.FormatInt(time.Now().Unix(), 10)
		})

		AfterEach(func() {
			azureClient.DeleteContainer(container.Name())
		})

		Context("when a file has some earlier versions", func() {
			It("restores to an earlier version, leaving snapshots soft deleted", func() {
				azureClient.WriteFileInContainer(container.Name(), fileName1, "TEST_BLOB_1_OLD")
				eTag1 = azureClient.WriteFileInContainer(container.Name(), fileName1, "TEST_BLOB_1")
				azureClient.WriteFileInContainer(container.Name(), fileName1, "TEST_BLOB_1_NEW")

				err := container.CopyBlobsFrom(container.Name(), []azure.BlobId{{Name: fileName1, ETag: eTag1}})

				Expect(err).NotTo(HaveOccurred())
				Expect(azureClient.ReadFileFromContainer(container.Name(), fileName1)).To(Equal("TEST_BLOB_1"))
			})
		})

		Context("when file is deleted", func() {
			It("restores it successfully", func() {
				eTag1 = azureClient.WriteFileInContainer(container.Name(), fileName1, "TEST_BLOB_1")
				azureClient.DeleteFileInContainer(container.Name(), fileName1)

				err := container.CopyBlobsFrom(container.Name(), []azure.BlobId{{Name: fileName1, ETag: eTag1}})

				Expect(err).NotTo(HaveOccurred())
				Expect(azureClient.ReadFileFromContainer(container.Name(), fileName1)).To(Equal("TEST_BLOB_1"))
			})
		})

		Context("when the source blob lives in a different container", func() {
			var differentContainer azure.Container

			BeforeEach(func() {
				containerName := azureClient.CreateContainerWithUniqueName("sdk-azure-test-")
				differentContainer = newContainer(containerName)
			})

			AfterEach(func() {
				azureClient.DeleteContainer(differentContainer.Name())
			})

			It("copies the blob to the destination container", func() {
				azureClient.WriteFileInContainer(container.Name(), fileName1, "TEST_BLOB_1_OLD")
				eTag1 = azureClient.WriteFileInContainer(container.Name(), fileName1, "TEST_BLOB_1")
				azureClient.WriteFileInContainer(container.Name(), fileName1, "TEST_BLOB_1_NEW")

				err := differentContainer.CopyBlobsFrom(container.Name(), []azure.BlobId{{Name: fileName1, ETag: eTag1}})

				Expect(err).NotTo(HaveOccurred())
				Expect(azureClient.ReadFileFromContainer(differentContainer.Name(), fileName1)).To(Equal("TEST_BLOB_1"))
			})
		})

		Context("when there is no matching snapshot", func() {
			It("returns an error", func() {
				err := container.CopyBlobsFrom(container.Name(), []azure.BlobId{{Name: fileName1, ETag: "wrong_eTag"}})

				Expect(err).To(MatchError(fmt.Sprintf("no \"%s\" blob with \"%s\" ETag found in container \"%s\"", fileName1, "wrong_eTag", container.Name())))
			})
		})

		Context("when the container is not reachable", func() {
			It("returns an error", func() {
				err := container.CopyBlobsFrom("wrong_container", []azure.BlobId{{}})

				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("CopyBlobsFromDifferentStorageAccount", func() {
		var differentAzureClient AzureClient
		var differentContainerName string
		var differentStorageAccount azure.StorageAccount

		BeforeEach(func() {
			containerName := azureClient.CreateContainerWithUniqueName("sdk-azure-test-")
			container = newContainer(containerName)

			differentStorageAccount = azure.StorageAccount{
				Name: MustHaveEnv("AZURE_DIFFERENT_STORAGE_ACCOUNT"),
				Key:  MustHaveEnv("AZURE_DIFFERENT_STORAGE_KEY"),
			}
			differentAzureClient = NewAzureClient(MustHaveEnv("AZURE_DIFFERENT_STORAGE_ACCOUNT"), MustHaveEnv("AZURE_DIFFERENT_STORAGE_KEY"))
			differentContainerName = differentAzureClient.CreateContainerWithUniqueName("sdk-azure-test-")
		})

		AfterEach(func() {
			azureClient.DeleteContainer(container.Name())
			differentAzureClient.DeleteContainer(differentContainerName)
		})

		It("copies the blob to the destination container", func() {
			differentAzureClient.WriteFileInContainer(differentContainerName, "a_file", "TEST_BLOB_1_OLD")
			eTag1 = differentAzureClient.WriteFileInContainer(differentContainerName, "a_file", "TEST_BLOB_1")
			differentAzureClient.WriteFileInContainer(differentContainerName, "a_file", "TEST_BLOB_1_NEW")

			err := container.CopyBlobsFromDifferentStorageAccount(differentStorageAccount, differentContainerName, []azure.BlobId{{Name: "a_file", ETag: eTag1}})

			Expect(err).NotTo(HaveOccurred())
			Expect(azureClient.ReadFileFromContainer(container.Name(), "a_file")).To(Equal("TEST_BLOB_1"))
		})
	})
})

func newContainer(containerName string) azure.Container {
	container, err := azure.NewSDKContainer(
		containerName,
		MustHaveEnv("AZURE_STORAGE_ACCOUNT"),
		MustHaveEnv("AZURE_STORAGE_KEY"),
		azure.DefaultEnvironment,
	)
	Expect(err).NotTo(HaveOccurred())
	return container
}
