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
	Describe("ListBlobs", func() {
		Context("when the backup succeeds", func() {
			var containerName = "bbr-test-azure-container"

			var fileName1, fileName2, fileName3 string

			BeforeEach(func() {
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
				container, err := azure.NewContainer(
					containerName,
					os.Getenv("AZURE_ACCOUNT_NAME"),
					os.Getenv("AZURE_ACCOUNT_KEY"),
				)
				Expect(err).NotTo(HaveOccurred())

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
