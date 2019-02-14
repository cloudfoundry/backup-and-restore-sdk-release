package incremental

import (
	"fmt"

	"github.com/cloudfoundry-incubator/bosh-backup-and-restore/executor"
)

type Restorer struct {
	liveBucket   Bucket
	backupBucket Bucket
}

func NewRestorer(liveBucket, backupBucket Bucket) Restorer {
	return Restorer{
		liveBucket:   liveBucket,
		backupBucket: backupBucket,
	}
}

func (p Restorer) Restore(bucketBackup BucketBackup) error {
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
