package helpers

import (
	"bytes"
	"io/ioutil"
	"os/exec"

	"encoding/json"

	"strings"

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
		"--name", blobName,
		"--delete-snapshots", "include")
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

func runAzureCommandSuccessfully(args ...string) *bytes.Buffer {
	outputBuffer := new(bytes.Buffer)
	errorBuffer := new(bytes.Buffer)

	azureStorageAccount := MustHaveEnv("AZURE_STORAGE_ACCOUNT")
	azureStorageKey := MustHaveEnv("AZURE_STORAGE_KEY")

	azureConfigDir, err := ioutil.TempDir("", "azure_")
	Expect(err).NotTo(HaveOccurred())

	azCmd := exec.Command("az", args...)
	azCmd.Stdout = outputBuffer
	azCmd.Stderr = errorBuffer
	azCmd.Env = append(azCmd.Env,
		"AZURE_STORAGE_ACCOUNT="+azureStorageAccount,
		"AZURE_STORAGE_KEY="+azureStorageKey,
		"AZURE_CONFIG_DIR="+azureConfigDir,
	)

	err = azCmd.Run()
	Expect(err).ToNot(HaveOccurred(), errorBuffer.String())

	return outputBuffer
}
