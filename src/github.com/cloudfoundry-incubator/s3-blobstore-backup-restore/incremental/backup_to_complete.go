package incremental

import "github.com/cloudfoundry-incubator/bosh-backup-and-restore/executor"

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
