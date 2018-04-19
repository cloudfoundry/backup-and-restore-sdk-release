package helpers

import (
	"bytes"
	"io/ioutil"
	"os/exec"

	. "github.com/onsi/gomega"
)

func DeleteFileInContainer(container, blobName string) {
	runAzureCommandOnContainerSuccessfully(
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

	runAzureCommandOnContainerSuccessfully(
		"storage",
		"blob",
		"upload",
		"--container-name", container,
		"--name", blobName,
		"--file", bodyFile.Name())
}

func runAzureCommandOnContainerSuccessfully(args ...string) *bytes.Buffer {
	outputBuffer := new(bytes.Buffer)
	errorBuffer := new(bytes.Buffer)

	azCmd := exec.Command("az", args...)
	azCmd.Stdout = outputBuffer
	azCmd.Stderr = errorBuffer

	err := azCmd.Run()
	Expect(err).ToNot(HaveOccurred(), errorBuffer.String())

	return outputBuffer
}
