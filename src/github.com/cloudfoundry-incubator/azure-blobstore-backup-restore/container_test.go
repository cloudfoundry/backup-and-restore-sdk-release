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
	var containerName = "bbr-test-azure-container"

	Describe("NewContainer", func() {
		It("builds a new Container", func() {
			container, err := azure.NewContainer(
				containerName,
				os.Getenv("AZURE_ACCOUNT_NAME"),
				os.Getenv("AZURE_ACCOUNT_KEY"),
			)

			Expect(err).NotTo(HaveOccurred())
			Expect(container.Name()).To(Equal(containerName))
		})

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

				Expect(err).To(MatchError(ContainSubstring("invalid credentials:")))
				Expect(container).To(Equal(azure.SDKContainer{}))
			})
		})
	})

	Describe("ListBlobs", func() {
		Context("when the backup succeeds", func() {
			var container azure.Container
			var fileName1, fileName2, fileName3 string

			BeforeEach(func() {
				var err error
				container, err = azure.NewContainer(
					containerName,
					os.Getenv("AZURE_ACCOUNT_NAME"),
					os.Getenv("AZURE_ACCOUNT_KEY"),
				)
				Expect(err).NotTo(HaveOccurred())

				fileName1 = "test_file_1_" + strconv.FormatInt(time.Now().Unix(), 10)
				fileName2 = "test_file_2_" + strconv.FormatInt(time.Now().Unix(), 10)
				fileName3 = "test_file_3_" + strconv.FormatInt(time.Now().Unix(), 10)
			})

			AfterEach(func() {
				deleteFileInContainer(containerName, fileName1)
				deleteFileInContainer(containerName, fileName2)
				deleteFileInContainer(containerName, fileName3)
			})

			It("returns a list of containers with files and hashes", func() {
				writeFileInContainer(containerName, fileName1, "TEST_BLOB_1_OLD")
				writeFileInContainer(containerName, fileName1, "TEST_BLOB_1")
				writeFileInContainer(containerName, fileName2, "TEST_BLOB_2_OLDEST")
				writeFileInContainer(containerName, fileName2, "TEST_BLOB_2_OLD")
				writeFileInContainer(containerName, fileName2, "TEST_BLOB_2")
				writeFileInContainer(containerName, fileName3, "TEST_BLOB_3")

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
				container, err := azure.NewContainer("NON-EXISTENT_CONTAINER", "", "")

				_, err = container.ListBlobs()

				Expect(err.Error()).To(ContainSubstring("failed listing blobs in container 'NON-EXISTENT_CONTAINER':"))
			})
		})
	})
})

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

	argsWithCredentials := append(args,
		"--account-name", os.Getenv("AZURE_ACCOUNT_NAME"),
		"--account-key", os.Getenv("AZURE_ACCOUNT_KEY"),
	)

	azCmd := exec.Command("az", argsWithCredentials...)
	azCmd.Stdout = outputBuffer
	azCmd.Stderr = errorBuffer

	err := azCmd.Run()
	Expect(err).ToNot(HaveOccurred(), errorBuffer.String())

	return outputBuffer
}
