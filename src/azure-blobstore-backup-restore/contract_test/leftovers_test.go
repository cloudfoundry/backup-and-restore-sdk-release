package contract_test

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Containers made by CreateContainerWithUniqueName("sdk-azure-test-").
var leftoverContainerName = regexp.MustCompile(`^sdk-azure-test-(\d{19})$`)

// Older than any run of this suite, so a sweep never touches a run in
// progress.
const leftoverMinAge = 3 * time.Hour

// deleteLeftoverContainers deletes the containers that earlier runs created
// in the storage account but didn't delete, because they were killed before
// their AfterEach ran. It only logs failures: a failed sweep shouldn't fail
// the suite.
func deleteLeftoverContainers(storageAccount, storageKey, endpointSuffix string) {
	output, err := tryAzureCommand(storageAccount, storageKey, endpointSuffix,
		"storage", "container", "list",
		"--prefix", "sdk-azure-test-",
		"--query", "[].name",
		"--output", "tsv")
	if err != nil {
		fmt.Fprintf(os.Stdout, "leftovers: listing containers in %s failed: %s\n", storageAccount, err)
		return
	}

	for _, containerName := range strings.Fields(output) {
		if !isLeftover(containerName) {
			continue
		}

		fmt.Fprintf(os.Stdout, "leftovers: deleting %s/%s\n", storageAccount, containerName)
		if _, err := tryAzureCommand(storageAccount, storageKey, endpointSuffix,
			"storage", "container", "delete", "--name", containerName); err != nil {
			fmt.Fprintf(os.Stdout, "leftovers: deleting %s/%s failed: %s\n", storageAccount, containerName, err)
		}
	}
}

func isLeftover(containerName string) bool {
	match := leftoverContainerName.FindStringSubmatch(containerName)
	if match == nil {
		return false
	}

	createdAt, err := strconv.ParseInt(match[1], 10, 64)
	if err != nil {
		return false
	}

	return time.Since(time.Unix(0, createdAt)) > leftoverMinAge
}

// tryAzureCommand runs az as AzureClient does, but returns errors rather
// than panicking.
func tryAzureCommand(storageAccount, storageKey, endpointSuffix string, args ...string) (string, error) {
	azureConfigDir, err := os.MkdirTemp("", "azure_")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(azureConfigDir)

	azCmd := exec.Command("az", args...)
	// Not os.Environ(): the task also sets AZURE_STORAGE_ACCOUNT and
	// AZURE_STORAGE_KEY, which az would use instead of the connection string.
	azCmd.Env = append(azCmd.Env,
		fmt.Sprintf(
			"AZURE_STORAGE_CONNECTION_STRING=DefaultEndpointsProtocol=https;AccountName=%s;AccountKey=%s;EndpointSuffix=%s;",
			storageAccount, storageKey, endpointSuffix,
		),
		"AZURE_CONFIG_DIR="+azureConfigDir,
	)

	output, err := azCmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("az %s: %w: %s", strings.Join(args, " "), err, exitErr.Stderr)
		}
		return "", fmt.Errorf("az %s: %w", strings.Join(args, " "), err)
	}
	return string(output), nil
}
