package incremental

import (
	"fmt"

	"strings"

	"errors"

	"github.com/cloudfoundry-incubator/bosh-backup-and-restore/executor"
)

type RestoreBucketPair struct {
	destinationLiveBucket Bucket
	sourceBackupBucket    Bucket
	executionStrategy     executor.ParallelExecutor
}

func NewRestoreBucketPair(liveBucket, backupBucket Bucket) RestoreBucketPair {
	exe := executor.NewParallelExecutor()
	exe.SetMaxInFlight(200)
	return RestoreBucketPair{
		destinationLiveBucket: liveBucket,
		sourceBackupBucket:    backupBucket,
		executionStrategy:     exe,
	}
}

type ExecutableBackup struct {
	file         string
	backupAction func(string) error
}

func (e ExecutableBackup) Execute() error {
	return e.backupAction(e.file)
}

func (p RestoreBucketPair) Restore(bucketBackup BucketBackup) error {
	var executables []executor.Executable
	for _, blob := range bucketBackup.Blobs {
		backedUpBlob := BackedUpBlob{
			Path:                blob,
			BackupDirectoryPath: bucketBackup.BackupDirectoryPath,
		}
		executables = append(executables, ExecutableBackup{file: blob, backupAction: func(blobKey string) error {
			return p.destinationLiveBucket.CopyBlobFromBucket(
				p.sourceBackupBucket,
				backedUpBlob.Path,
				backedUpBlob.LiveBlobPath(),
			)
		}})
	}

	errs := p.executionStrategy.Run([][]executor.Executable{executables})
	if len(errs) != 0 {
		return formatErrors(
			fmt.Sprintf("failed to restore bucket %s", p.destinationLiveBucket.Name()),
			errs,
		)
	}

	return nil
}

func (p RestoreBucketPair) CheckValidity() error {
	if p.destinationLiveBucket.Name() == p.sourceBackupBucket.Name() {
		return errors.New("live bucket and backup bucket cannot be the same")
	}

	return nil
}

func formatErrors(contextString string, errors []error) error {
	errorStrings := make([]string, len(errors))
	for i, err := range errors {
		errorStrings[i] = err.Error()
	}
	return fmt.Errorf("%s:\n%s", contextString, strings.Join(errorStrings, "\n"))
}
