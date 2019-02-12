package incremental

import (
	"fmt"
	"strings"

	"github.com/cloudfoundry-incubator/bosh-backup-and-restore/executor"
)

type BackupCompleter struct {
	BackupsToComplete map[string]BackupToComplete
}

func (b BackupCompleter) Run() error {
	for _, backupToComplete := range b.BackupsToComplete {
		e := executor.NewParallelExecutor()
		e.SetMaxInFlight(200)

		errs := e.Run(backupToComplete.executables())
		if len(errs) != 0 {
			return formatExecutorErrors(fmt.Sprintf("failed to complete backup"), errs)
		}

		err := backupToComplete.BackupDirectory.MarkComplete()
		if err != nil {
			return fmt.Errorf("failed to complete backup: %s", err)
		}
	}

	return nil
}

type BackupToComplete struct {
	BackupBucket    Bucket
	BackupDirectory BackupDirectory
	BlobsToCopy     []BackedUpBlob
}

func (b BackupToComplete) executables() [][]executor.Executable {
	var executables []executor.Executable
	for _, blob := range b.BlobsToCopy {
		executable := copyBlobWithinBucketExecutable{
			bucket: b.BackupBucket,
			src:    blob.Path,
			dst:    joinBlobPath(b.BackupDirectory.Path, blob.LiveBlobPath()),
		}
		executables = append(executables, executable)
	}

	return [][]executor.Executable{executables}
}

type copyBlobWithinBucketExecutable struct {
	src    string
	dst    string
	bucket Bucket
}

func (e copyBlobWithinBucketExecutable) Execute() error {
	return e.bucket.CopyBlobWithinBucket(e.src, e.dst)
}

func formatExecutorErrors(context string, errors []error) error {
	messages := make([]string, len(errors))
	for i, err := range errors {
		messages[i] = err.Error()
	}

	return fmt.Errorf("%s: %s", context, strings.Join(messages, "\n"))
}
