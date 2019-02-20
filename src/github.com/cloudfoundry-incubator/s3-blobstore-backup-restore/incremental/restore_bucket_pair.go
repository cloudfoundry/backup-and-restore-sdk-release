package incremental

import (
	"fmt"

	"github.com/cloudfoundry-incubator/bosh-backup-and-restore/executor"
)

type RestoreBucketPair struct {
	ConfigLiveBucket     Bucket
	ArtifactBackupBucket Bucket
}

func (p RestoreBucketPair) Restore(backup Backup) error {
	var executables []executor.Executable
	for _, blob := range backup.Blobs {
		backedUpBlob := BackedUpBlob{
			Path:                blob,
			BackupDirectoryPath: backup.SrcBackupDirectoryPath,
		}
		executables = append(executables, copyBlobFromBucketExecutable{
			src:       backedUpBlob.Path,
			dst:       backedUpBlob.LiveBlobPath(),
			srcBucket: p.ArtifactBackupBucket,
			dstBucket: p.ConfigLiveBucket,
		})
	}

	e := executor.NewParallelExecutor()
	e.SetMaxInFlight(200)

	errs := e.Run([][]executor.Executable{executables})
	if len(errs) != 0 {
		return formatExecutorErrors(
			fmt.Sprintf("failed to restore bucket %s", p.ConfigLiveBucket.Name()),
			errs,
		)
	}

	return nil
}
