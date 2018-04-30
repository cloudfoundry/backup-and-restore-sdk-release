package azure_test

import (
	"strconv"
	"time"

	"github.com/cloudfoundry-incubator/azure-blobstore-backup-restore"
	. "github.com/cloudfoundry-incubator/azure-blobstore-backup-restore/system_tests/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Container", func() {
	Describe("NewContainer", func() {
		Context("when the account name is invalid", func() {
			It("returns an error", func() {
				container, err := azure.NewContainer("", "\n", "")

				Expect(err).To(MatchError("invalid account name: '\n'"))
				Expect(container).To(Equal(azure.SDKContainer{}))
			})
		})

		Context("when the account key is not valid base64", func() {
			It("returns an error", func() {
				container, err := azure.NewContainer("", "", "#")

				Expect(err).To(MatchError(ContainSubstring("invalid storage key: '")))
				Expect(container).To(Equal(azure.SDKContainer{}))
			})
		})
	})

	Describe("Name", func() {
		It("returns the container name", func() {
			name := "container-name"

			container, err := azure.NewContainer(name, "", "")

			Expect(err).NotTo(HaveOccurred())
			Expect(container.Name()).To(Equal(name))
		})
	})

	Describe("SoftDeleteEnabled", func() {
		Context("when soft delete is enabled on the container's storage service", func() {
			It("returns true", func() {
				container := newContainer()

				enabled, err := container.SoftDeleteEnabled()

				Expect(err).NotTo(HaveOccurred())
				Expect(enabled).To(BeTrue())

				DeleteContainer(container.Name())
			})
		})

		Context("when soft delete is disabled on the container's storage service", func() {
			It("returns false", func() {
				container, err := azure.NewContainer(
					"",
					MustHaveEnv("AZURE_STORAGE_ACCOUNT_NO_SOFT_DELETE"),
					MustHaveEnv("AZURE_STORAGE_KEY_NO_SOFT_DELETE"),
				)

				enabled, err := container.SoftDeleteEnabled()

				Expect(err).NotTo(HaveOccurred())
				Expect(enabled).To(BeFalse())
			})
		})

		Context("when retrieving the storage service properties fails", func() {
			It("returns an error", func() {
				container, err := azure.NewContainer("", "", "")

				_, err = container.SoftDeleteEnabled()

				Expect(err.Error()).To(ContainSubstring("failed fetching properties for storage account: '"))
			})
		})
	})

	Describe("ListBlobs", func() {
		Context("when the container has a few files and snapshots", func() {
			It("returns a list of containers with files and their etags", func() {
				// arrange
				container := newContainer()

				fileName1 := "test_file_1_" + strconv.FormatInt(time.Now().Unix(), 10)
				fileName2 := "test_file_2_" + strconv.FormatInt(time.Now().Unix(), 10)
				fileName3 := "test_file_3_" + strconv.FormatInt(time.Now().Unix(), 10)

				WriteFileInContainer(container.Name(), fileName1, "TEST_BLOB_1_OLD")
				etag1 := WriteFileInContainer(container.Name(), fileName1, "TEST_BLOB_1")
				WriteFileInContainer(container.Name(), fileName2, "TEST_BLOB_2_OLDEST")
				WriteFileInContainer(container.Name(), fileName2, "TEST_BLOB_2_OLD")
				etag2 := WriteFileInContainer(container.Name(), fileName2, "TEST_BLOB_2")
				etag3 := WriteFileInContainer(container.Name(), fileName3, "TEST_BLOB_3")

				// act
				blobs, err := container.ListBlobs()

				// assert
				Expect(err).NotTo(HaveOccurred())
				Expect(blobs).To(Equal([]azure.Blob{
					{Name: fileName1, Etag: etag1},
					{Name: fileName2, Etag: etag2},
					{Name: fileName3, Etag: etag3},
				}))

				// teardown
				DeleteContainer(container.Name())
			})
		})

		Context("when the container has a lots of files", func() {
			It("paginates correctly", func() {
				container, err := azure.NewContainer(
					MustHaveEnv("AZURE_CONTAINER_NAME_MANY_FILES"),
					MustHaveEnv("AZURE_STORAGE_ACCOUNT"),
					MustHaveEnv("AZURE_STORAGE_KEY"),
				)
				Expect(err).NotTo(HaveOccurred())

				blobs, err := container.ListBlobs()

				Expect(err).NotTo(HaveOccurred())
				Expect(len(blobs)).To(Equal(10104))
			})
		})

		Context("when listing the blobs fails", func() {
			It("returns an error", func() {
				container, err := azure.NewContainer(
					"NON-EXISTENT-CONTAINER",
					MustHaveEnv("AZURE_STORAGE_ACCOUNT"),
					MustHaveEnv("AZURE_STORAGE_KEY"),
				)

				_, err = container.ListBlobs()

				Expect(err.Error()).To(ContainSubstring("failed listing blobs in container 'NON-EXISTENT-CONTAINER':"))
			})
		})
	})
})

func newContainer() azure.Container {
	containerName := "sdk-azure-test-" + strconv.FormatInt(time.Now().UnixNano(), 10)

	CreateContainer(containerName)

	container, err := azure.NewContainer(
		containerName,
		MustHaveEnv("AZURE_STORAGE_ACCOUNT"),
		MustHaveEnv("AZURE_STORAGE_KEY"),
	)
	Expect(err).NotTo(HaveOccurred())

	return container
}
