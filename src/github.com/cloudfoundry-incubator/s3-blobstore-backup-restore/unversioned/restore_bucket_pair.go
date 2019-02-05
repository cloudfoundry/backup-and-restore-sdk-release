package unversioned

import (
	"fmt"

	"github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/incremental"

	"errors"

	"github.com/cloudfoundry-incubator/bosh-backup-and-restore/executor"
)

type RestoreBucketPair struct {
	destinationLiveBucket incremental.Bucket
	sourceBackupBucket    incremental.Bucket
	executionStrategy     executor.ParallelExecutor
}

func NewRestoreBucketPair(liveBucket, backupBucket incremental.Bucket) RestoreBucketPair {
	exe := executor.NewParallelExecutor()
	exe.SetMaxInFlight(200)
	return RestoreBucketPair{
		destinationLiveBucket: liveBucket,
		sourceBackupBucket:    backupBucket,
		executionStrategy:     exe,
	}
}

func (p RestoreBucketPair) Restore(bucketBackup incremental.BucketBackup) error {
	var executables []executor.Executable
	for _, blob := range bucketBackup.Blobs {
		backedUpBlob := incremental.BackedUpBlob{
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
