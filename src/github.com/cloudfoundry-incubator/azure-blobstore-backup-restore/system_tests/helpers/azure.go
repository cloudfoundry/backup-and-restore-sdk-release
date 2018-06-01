package helpers

import (
	"bytes"
	"io/ioutil"
	"os/exec"

	"encoding/json"

	"strings"

	. "github.com/onsi/gomega"
)

type AzureClient struct {
	storageAccount string
	storageKey     string
}

func NewAzureClient(storageAccount, storageKey string) AzureClient {
	return AzureClient{storageAccount: storageAccount, storageKey: storageKey}
}

func (c AzureClient) DeleteContainer(name string) {
	c.runAzureCommandSuccessfully(
		"storage",
		"container",
		"delete",
		"--name",
		name,
	)
}

func (c AzureClient) DeleteFileInContainer(container, blobName string) {
	c.runAzureCommandSuccessfully(
		"storage",
		"blob",
		"delete",
		"--container-name", container,
		"--name", blobName,
		"--delete-snapshots", "include")
}

func (c AzureClient) WriteFileInContainer(container, blobName, body string) string {
	bodyFile, _ := ioutil.TempFile("", "write_file_in_container_")
	bodyFile.WriteString(body)
	bodyFile.Close()

	outputBuffer := c.runAzureCommandSuccessfully(
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

func (c AzureClient) CreateContainer(name string) {
	c.runAzureCommandSuccessfully(
		"storage",
		"container",
		"create",
		"--name",
		name,
		"--fail-on-exist",
	)
}

func (c AzureClient) ReadFileFromContainer(container, blobName string) string {
	bodyFile, err := ioutil.TempFile("", "read_file_from_container_")
	Expect(err).NotTo(HaveOccurred())

	c.runAzureCommandSuccessfully(
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

func (c AzureClient) runAzureCommandSuccessfully(args ...string) *bytes.Buffer {
	outputBuffer := new(bytes.Buffer)
	errorBuffer := new(bytes.Buffer)

	azureConfigDir, err := ioutil.TempDir("", "azure_")
	Expect(err).NotTo(HaveOccurred())

	azCmd := exec.Command("az", args...)
	azCmd.Stdout = outputBuffer
	azCmd.Stderr = errorBuffer
	azCmd.Env = append(azCmd.Env,
		"AZURE_STORAGE_ACCOUNT="+c.storageAccount,
		"AZURE_STORAGE_KEY="+c.storageKey,
		"AZURE_CONFIG_DIR="+azureConfigDir,
	)

	err = azCmd.Run()
	Expect(err).ToNot(HaveOccurred(), errorBuffer.String())

	return outputBuffer
}
