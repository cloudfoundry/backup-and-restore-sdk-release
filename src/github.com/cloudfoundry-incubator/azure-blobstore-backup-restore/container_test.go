package azure_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"

	"strconv"
	"time"

	"github.com/cloudfoundry-incubator/azure-blobstore-backup-restore"
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
			})
		})

		Context("when soft delete is disabled on the container's storage service", func() {
			It("returns false", func() {
				container, err := azure.NewContainer(
					mustHaveEnv("AZURE_CONTAINER_NAME_NO_SOFT_DELETE"),
					mustHaveEnv("AZURE_STORAGE_ACCOUNT_NO_SOFT_DELETE"),
					mustHaveEnv("AZURE_STORAGE_KEY_NO_SOFT_DELETE"),
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
		Context("when the backup succeeds", func() {
			var container azure.Container
			var fileName1, fileName2, fileName3 string

			BeforeEach(func() {
				container = newContainer()

				deleteAllBlobsInContainer(container.Name())

				fileName1 = "test_file_1_" + strconv.FormatInt(time.Now().Unix(), 10)
				fileName2 = "test_file_2_" + strconv.FormatInt(time.Now().Unix(), 10)
				fileName3 = "test_file_3_" + strconv.FormatInt(time.Now().Unix(), 10)
			})

			AfterEach(func() {
				deleteFileInContainer(container.Name(), fileName1)
				deleteFileInContainer(container.Name(), fileName2)
				deleteFileInContainer(container.Name(), fileName3)
			})

			It("returns a list of containers with files and hashes", func() {
				writeFileInContainer(container.Name(), fileName1, "TEST_BLOB_1_OLD")
				writeFileInContainer(container.Name(), fileName1, "TEST_BLOB_1")
				writeFileInContainer(container.Name(), fileName2, "TEST_BLOB_2_OLDEST")
				writeFileInContainer(container.Name(), fileName2, "TEST_BLOB_2_OLD")
				writeFileInContainer(container.Name(), fileName2, "TEST_BLOB_2")
				writeFileInContainer(container.Name(), fileName3, "TEST_BLOB_3")

				blobs, err := container.ListBlobs()

				Expect(err).NotTo(HaveOccurred())
				Expect(blobs).To(Equal([]azure.Blob{
					{Name: fileName1, Hash: "R1M39xrrgP7eS+jJHBWu1A=="},
					{Name: fileName2, Hash: "L+IcKub+0Og4CXjKqA1/3w=="},
					{Name: fileName3, Hash: "7VBVkm19ll+P6THGtqGHww=="},
				}))
			})
		})

		Context("when listing the blobs fails", func() {
			It("returns an error", func() {
				container, err := azure.NewContainer(
					"NON-EXISTENT-CONTAINER",
					mustHaveEnv("AZURE_STORAGE_ACCOUNT"),
					mustHaveEnv("AZURE_STORAGE_KEY"),
				)

				_, err = container.ListBlobs()

				Expect(err.Error()).To(ContainSubstring("failed listing blobs in container 'NON-EXISTENT-CONTAINER':"))
			})
		})
	})
})

func containerName() string {
	return mustHaveEnv("AZURE_CONTAINER_NAME")
}

func newContainer() azure.Container {
	container, err := azure.NewContainer(
		containerName(),
		mustHaveEnv("AZURE_STORAGE_ACCOUNT"),
		mustHaveEnv("AZURE_STORAGE_KEY"),
	)
	Expect(err).NotTo(HaveOccurred())
	return container
}

func deleteAllBlobsInContainer(containerName string) {
	runAzureCommandSuccessfully(
		"storage",
		"blob",
		"delete-batch",
		"--source",
		"bbr-test-azure-container",
		"--if-match",
		"*",
	)
}

func deleteFileInContainer(containerName, blobName string) {
	runAzureCommandSuccessfully(
		"storage",
		"blob",
		"delete",
		"--container-name", containerName,
		"--name", blobName)
}

func writeFileInContainer(containerName, blobName, body string) {
	bodyFile, _ := ioutil.TempFile("", "")
	bodyFile.WriteString(body)
	bodyFile.Close()

	runAzureCommandSuccessfully(
		"storage",
		"blob",
		"upload",
		"--container-name", containerName,
		"--name", blobName,
		"--file", bodyFile.Name())
}

func runAzureCommandSuccessfully(args ...string) *bytes.Buffer {
	outputBuffer := new(bytes.Buffer)
	errorBuffer := new(bytes.Buffer)

	mustHaveEnv("AZURE_STORAGE_ACCOUNT")
	mustHaveEnv("AZURE_STORAGE_KEY")

	azCmd := exec.Command("az", args...)
	azCmd.Stdout = outputBuffer
	azCmd.Stderr = errorBuffer

	err := azCmd.Run()
	Expect(err).ToNot(HaveOccurred(), errorBuffer.String())

	return outputBuffer
}

func mustHaveEnv(keyname string) string {
	val := os.Getenv(keyname)
	Expect(val).NotTo(BeEmpty(), "Need "+keyname+" for the test")
	return val
}
