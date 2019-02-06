package incremental

import (
	"fmt"

	"strings"

	"errors"

	"github.com/cloudfoundry-incubator/bosh-backup-and-restore/executor"
)

//go:generate counterfeiter -o fakes/fake_restore_bucket_pair.go . RestoreBucketPair
type RestoreBucketPair interface {
	CheckValidity() error
	Restore(bucketBackup BucketBackup) error
}

type IncrementalBucketPair struct {
	liveBucket        Bucket
	backupBucket      Bucket
	executionStrategy executor.ParallelExecutor
}

func NewIncrementalBucketPair(liveBucket, backupBucket Bucket) IncrementalBucketPair {
	exe := executor.NewParallelExecutor()
	exe.SetMaxInFlight(200)
	return IncrementalBucketPair{
		liveBucket:        liveBucket,
		backupBucket:      backupBucket,
		executionStrategy: exe,
	}
}

type ExecutableBackup struct {
	file         string
	backupAction func(string) error
}

func (e ExecutableBackup) Execute() error {
	return e.backupAction(e.file)
}

func (p IncrementalBucketPair) Restore(bucketBackup BucketBackup) error {
	var executables []executor.Executable
	for _, blob := range bucketBackup.Blobs {
		executables = append(executables, ExecutableBackup{file: blob, backupAction: func(blobKey string) error {
			// NEEDS WORK
			//return p.liveBucket.CopyBlobFromBucket(
			//	bucketBackup.BackupDirectoryPath + blobKey,
			//	"",
			//	bucketBackup.BucketRegion,
			//	)
			//)
			return nil
		}})
	}

	errs := p.executionStrategy.Run([][]executor.Executable{executables})
	if len(errs) != 0 {
		return formatErrors(
			fmt.Sprintf("failed to restore bucket %s", p.liveBucket.Name()),
			errs,
		)
	}

	return nil
}

func (p IncrementalBucketPair) CheckValidity() error {
	if p.liveBucket.Name() == p.backupBucket.Name() {
		return errors.New("live bucket and backup bucket cannot be the same")
	}

	return nil
}

func formatErrors(contextString string, errors []error) error {
	errorStrings := make([]string, len(errors))
	for i, err := range errors {
		errorStrings[i] = err.Error()
	}
	return fmt.Errorf("%s: %s", contextString, strings.Join(errorStrings, "\n"))
}
