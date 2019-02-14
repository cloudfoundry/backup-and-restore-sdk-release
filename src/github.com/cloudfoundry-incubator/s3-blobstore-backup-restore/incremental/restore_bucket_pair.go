package incremental

import (
	"fmt"

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

func (p RestoreBucketPair) Restore(bucketBackup BucketBackup) error {
	var executables []executor.Executable
	for _, blob := range bucketBackup.Blobs {
		backedUpBlob := BackedUpBlob{
			Path:                blob,
			BackupDirectoryPath: bucketBackup.BackupDirectoryPath,
		}
		executables = append(executables, copyBlobFromBucketExecutable{
			src:       backedUpBlob.Path,
			dst:       backedUpBlob.LiveBlobPath(),
			srcBucket: p.sourceBackupBucket,
			dstBucket: p.destinationLiveBucket,
		})
	}

	errs := p.executionStrategy.Run([][]executor.Executable{executables})
	if len(errs) != 0 {
		return formatExecutorErrors(
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
