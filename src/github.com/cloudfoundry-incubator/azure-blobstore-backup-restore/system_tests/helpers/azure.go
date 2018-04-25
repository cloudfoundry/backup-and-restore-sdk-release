package helpers

import (
	"bytes"
	"io/ioutil"
	"os/exec"

	"strconv"
	"time"

	. "github.com/onsi/gomega"
)

func DeleteContainer(name string) {
	runAzureCommandSuccessfully(
		"storage",
		"container",
		"delete",
		"--name",
		name,
	)
}

func DeleteFileInContainer(container, blobName string) {
	runAzureCommandSuccessfully(
		"storage",
		"blob",
		"delete",
		"--container-name", container,
		"--name", blobName)
}

func WriteFileInContainer(container, blobName, body string) {
	bodyFile, _ := ioutil.TempFile("", "")
	bodyFile.WriteString(body)
	bodyFile.Close()

	runAzureCommandSuccessfully(
		"storage",
		"blob",
		"upload",
		"--container-name", container,
		"--name", blobName,
		"--file", bodyFile.Name())
}

func ContainerName() string {
	return "sdk-azure-test-" + strconv.FormatInt(time.Now().UnixNano(), 10)
}

func CreateContainer(name string) {
	runAzureCommandSuccessfully(
		"storage",
		"container",
		"create",
		"--name",
		name,
		"--fail-on-exist",
	)
}

func runAzureCommandSuccessfully(args ...string) *bytes.Buffer {
	outputBuffer := new(bytes.Buffer)
	errorBuffer := new(bytes.Buffer)

	MustHaveEnv("AZURE_STORAGE_ACCOUNT")
	MustHaveEnv("AZURE_STORAGE_KEY")

	azCmd := exec.Command("az", args...)
	azCmd.Stdout = outputBuffer
	azCmd.Stderr = errorBuffer

	err := azCmd.Run()
	Expect(err).ToNot(HaveOccurred(), errorBuffer.String())

	return outputBuffer
}
