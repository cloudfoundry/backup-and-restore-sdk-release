package s3bucket_test

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Buckets made by setUpUnversionedBucket and setUpVersionedBucket.
var leftoverBucketName = regexp.MustCompile(`^sdk-integration-test-(\d{19})$`)

// Older than any run of this suite, so a sweep never touches a run in
// progress.
const leftoverMinAge = 3 * time.Hour

// deleteLeftoverBuckets deletes the buckets that earlier runs created but
// didn't delete, because they were killed before their AfterEach ran. It only
// logs failures: a failed sweep shouldn't fail the suite.
func deleteLeftoverBuckets(endpoint string) {
	output, err := tryAwsCommand(endpoint, "s3api", "list-buckets",
		"--query", "Buckets[?starts_with(Name, 'sdk-integration-test-')].Name",
		"--output", "text")
	if err != nil {
		fmt.Fprintf(os.Stdout, "leftovers: listing buckets failed: %s\n", err)
		return
	}

	for _, bucketName := range strings.Fields(output) {
		if !isLeftover(bucketName) {
			continue
		}

		fmt.Fprintf(os.Stdout, "leftovers: deleting s3://%s\n", bucketName)
		if err := deleteLeftoverBucket(bucketName, endpoint); err != nil {
			fmt.Fprintf(os.Stdout, "leftovers: deleting s3://%s failed: %s\n", bucketName, err)
		}
	}
}

// deleteLeftoverBucket deletes every object version and delete marker first:
// `aws s3 rb --force` leaves them, and a versioned bucket can't be deleted
// until they're gone.
func deleteLeftoverBucket(bucketName, endpoint string) error {
	output, err := tryAwsCommand(endpoint, "s3api", "list-object-versions", "--bucket", bucketName)
	if err != nil {
		return err
	}

	var response VersionsResponse
	if strings.TrimSpace(output) != "" {
		if err := json.Unmarshal([]byte(output), &response); err != nil {
			return err
		}
	}

	for _, version := range append(response.Versions, response.DeleteMarkers...) {
		if _, err := tryAwsCommand(endpoint, "s3api", "delete-object",
			"--bucket", bucketName,
			"--key", version.Key,
			"--version-id", version.VersionId); err != nil {
			return err
		}
	}

	_, err = tryAwsCommand(endpoint, "s3", "rb", "s3://"+bucketName, "--force")
	return err
}

func isLeftover(bucketName string) bool {
	match := leftoverBucketName.FindStringSubmatch(bucketName)
	if match == nil {
		return false
	}

	createdAt, err := strconv.ParseInt(match[1], 10, 64)
	if err != nil {
		return false
	}

	return time.Since(time.Unix(0, createdAt)) > leftoverMinAge
}

// tryAwsCommand is runAwsCommand without the assertions.
func tryAwsCommand(endpoint string, args ...string) (string, error) {
	awsCmd := newAwsCommand(append(constructBaseCmd(endpoint), args...))
	output, err := awsCmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("aws %s: %w: %s", strings.Join(args, " "), err, exitErr.Stderr)
		}
		return "", fmt.Errorf("aws %s: %w", strings.Join(args, " "), err)
	}
	return string(output), nil
}
