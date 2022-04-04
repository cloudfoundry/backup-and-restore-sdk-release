package incremental

import (
	"github.com/cloudfoundry/backup-and-restore-sdk-release/src/s3-blobstore-backup-restore/blobpath"
)

type BackedUpBlob struct {
	Path                string
	BackupDirectoryPath string
}

func (b BackedUpBlob) LiveBlobPath() string {
	return blobpath.TrimPrefix(b.Path, b.BackupDirectoryPath)
}

func joinBlobPath(prefix, suffix string) string {
	return blobpath.Join(prefix, suffix)
}
