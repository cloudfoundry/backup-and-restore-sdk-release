package incremental

import "github.com/cloudfoundry-incubator/bosh-backup-and-restore/executor"

type BackupBucketPair struct {
	ConfigLiveBucket   Bucket
	ConfigBackupBucket Bucket
}

func (b BackupBucketPair) CopyNewLiveBlobsToBackup(backedUpBlobs []BackedUpBlob, liveBlobs []Blob, backupDirPath string) ([]BackedUpBlob, error) {
	backedUpBlobsMap := make(map[string]BackedUpBlob)
	for _, backedUpBlob := range backedUpBlobs {
		backedUpBlobsMap[backedUpBlob.LiveBlobPath()] = backedUpBlob
	}

	var (
		executables   []executor.Executable
		existingBlobs []BackedUpBlob
	)
	for _, liveBlob := range liveBlobs {
		backedUpBlob, exists := backedUpBlobsMap[liveBlob.Path()]

		if !exists {
			executable := copyBlobFromBucketExecutable{
				src:       liveBlob.Path(),
				dst:       joinBlobPath(backupDirPath, liveBlob.Path()),
				srcBucket: b.ConfigLiveBucket,
				dstBucket: b.ConfigBackupBucket,
			}
			executables = append(executables, executable)
		} else {
			existingBlobs = append(existingBlobs, backedUpBlob)
		}
	}

	e := executor.NewParallelExecutor()
	e.SetMaxInFlight(200)

	errs := e.Run([][]executor.Executable{executables})
	if len(errs) != 0 {
		return nil, formatExecutorErrors("failing copying blobs in parallel", errs)
	}

	return existingBlobs, nil
}

type copyBlobFromBucketExecutable struct {
	src       string
	dst       string
	dstBucket Bucket
	srcBucket Bucket
}

func (e copyBlobFromBucketExecutable) Execute() error {
	return e.dstBucket.CopyBlobFromBucket(e.srcBucket, e.src, e.dst)
}
