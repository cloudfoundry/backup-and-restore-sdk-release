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

// Buckets made by CreateBucketWithTimestampedName. The project is shared
// with other pipelines, so the pattern is as narrow as the names allow.
var leftoverBucketName = regexp.MustCompile(`^contract-test-[a-z_]+-(\d{19})$`)

// Older than any run of this suite, so a sweep never touches a run in
// progress.
const leftoverMinAge = 3 * time.Hour

// deleteLeftoverBuckets deletes the buckets that earlier runs created but
// didn't delete, because they were killed before their AfterEach ran. It only
// logs failures: a failed sweep shouldn't fail the suite.
func deleteLeftoverBuckets(projectName string) {
	output, err := exec.Command("gcloud", "storage", "buckets", "list",
		"--project", projectName,
		"--filter", "name ~ ^contract-test-",
		"--format", "value(name)",
	).Output()
	if err != nil {
		fmt.Fprintf(os.Stdout, "leftovers: listing buckets failed: %s\n", err)
		return
	}

	for _, bucketName := range strings.Fields(string(output)) {
		if !isLeftover(bucketName) {
			continue
		}

		fmt.Fprintf(os.Stdout, "leftovers: deleting gs://%s\n", bucketName)
		if output, err := exec.Command("gsutil", "rm", "-r", "gs://"+bucketName).CombinedOutput(); err != nil {
			fmt.Fprintf(os.Stdout, "leftovers: deleting gs://%s failed: %s\n%s\n", bucketName, err, output)
		}
	}
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
