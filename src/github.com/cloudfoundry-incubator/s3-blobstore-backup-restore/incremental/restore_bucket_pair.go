package incremental

import (
	"fmt"

	"errors"

	"github.com/cloudfoundry-incubator/bosh-backup-and-restore/executor"
)

type RestoreBucketPair struct {
	liveBucket   Bucket
	backupBucket Bucket
}

func NewRestoreBucketPair(liveBucket, backupBucket Bucket) RestoreBucketPair {
	return RestoreBucketPair{
		liveBucket:   liveBucket,
		backupBucket: backupBucket,
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
			srcBucket: p.backupBucket,
			dstBucket: p.liveBucket,
		})
	}

	e := executor.NewParallelExecutor()
	e.SetMaxInFlight(200)

	errs := e.Run([][]executor.Executable{executables})
	if len(errs) != 0 {
		return formatExecutorErrors(
			fmt.Sprintf("failed to restore bucket %s", p.liveBucket.Name()),
			errs,
		)
	}

	return nil
}

func (p RestoreBucketPair) CheckValidity() error {
	if p.liveBucket.Name() == p.backupBucket.Name() {
		return errors.New("live bucket and backup bucket cannot be the same")
	}

	return nil
}
