package incremental

import (
	"fmt"
)

const backupComplete = "backup_complete"

type BackupDirectory struct {
	Path   string
	Bucket Bucket
}

func (b BackupDirectory) ListBlobs() ([]BackedUpBlob, error) {
	blobs, err := b.Bucket.ListBlobs(b.Path)
	if err != nil {
		return nil, fmt.Errorf("failed listing blobs in backup directory '%s': %s", b.Path, err)
	}

	var backedUpBlobs []BackedUpBlob
	for _, blob := range blobs {
		backedUpBlobs = append(backedUpBlobs, BackedUpBlob{
			Path:                blob.Path(),
			BackupDirectoryPath: b.Path,
		})
	}

	return backedUpBlobs, nil
}

func (b BackupDirectory) IsComplete() (bool, error) {
	hasBlob, err := b.Bucket.HasBlob(b.backupCompletePath())
	if err != nil {
		return false, fmt.Errorf("failed checking if backup directory '%s' is complete: %s", b.Path, err)
	}

	return hasBlob, nil
}

func (b BackupDirectory) MarkComplete() error {
	err := b.Bucket.UploadBlob(b.backupCompletePath(), "")
	if err != nil {
		return fmt.Errorf("failed marking backup directory '%s' as complete: %s", b.Path, err)
	}

	return nil
}

func (b BackupDirectory) backupCompletePath() string {
	return joinBlobPath(b.Path, backupComplete)
}
