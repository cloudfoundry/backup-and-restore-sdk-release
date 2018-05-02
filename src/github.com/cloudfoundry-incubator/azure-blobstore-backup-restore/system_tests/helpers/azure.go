package helpers

import (
	"bytes"
	"io/ioutil"
	"os/exec"

	"encoding/json"

	"strings"

	"strconv"

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

func WriteFileInContainer(container, blobName, body string) string {
	bodyFile, _ := ioutil.TempFile("", "write_file_in_container_")
	bodyFile.WriteString(body)
	bodyFile.Close()

	outputBuffer := runAzureCommandSuccessfully(
		"storage",
		"blob",
		"upload",
		"--container-name", container,
		"--name", blobName,
		"--file", bodyFile.Name())

	var output = make(map[string]string)
	json.Unmarshal(outputBuffer.Bytes(), &output)

	return strings.Trim(output["etag"], "\"")
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

func ReadFileFromContainer(container, blobName string) string {
	bodyFile, err := ioutil.TempFile("", "read_file_from_container_")
	Expect(err).NotTo(HaveOccurred())

	runAzureCommandSuccessfully(
		"storage",
		"blob",
		"download",
		"--container-name", container,
		"--name", blobName,
		"--file", bodyFile.Name())

	body, err := ioutil.ReadFile(bodyFile.Name())
	Expect(err).NotTo(HaveOccurred())

	return string(body)
}

func NumberOfUndeletedSnapshots(container string) int {
	outputBuffer := runAzureCommandSuccessfully(
		"storage",
		"blob",
		"list",
		"--container-name", container,
		"--include", "s",
		"--query", "length(@)")

	outputString := string(outputBuffer.Bytes())

	outputNumber, err := strconv.Atoi(strings.TrimSpace(outputString))
	Expect(err).NotTo(HaveOccurred())

	return outputNumber
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
