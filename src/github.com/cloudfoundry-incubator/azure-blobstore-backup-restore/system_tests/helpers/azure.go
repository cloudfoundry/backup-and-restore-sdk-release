package helpers

import (
	"bytes"
	"io/ioutil"
	"os/exec"

	. "github.com/onsi/gomega"
)

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

func runAzureCommandSuccessfully(args ...string) *bytes.Buffer {
	outputBuffer := new(bytes.Buffer)
	errorBuffer := new(bytes.Buffer)

	argsWithCredentials := append(args,
		"--account-name", MustHaveEnv("AZURE_ACCOUNT_NAME"),
		"--account-key", MustHaveEnv("AZURE_ACCOUNT_KEY"),
	)

	azCmd := exec.Command("az", argsWithCredentials...)
	azCmd.Stdout = outputBuffer
	azCmd.Stderr = errorBuffer

	err := azCmd.Run()
	Expect(err).ToNot(HaveOccurred(), errorBuffer.String())

	return outputBuffer
}
