package gcs

import (
	"fmt"
	"strings"
)

//go:generate counterfeiter -o fakes/fake_backup_finder.go . BackupFinder
type BackupFinder interface {
	ListBlobs() (map[string]Blob, error)
}

type LastBackupFinder struct {
	bucket Bucket
}

func NewLastBackupFinder(bucket Bucket) *LastBackupFinder {
	return &LastBackupFinder{
		bucket: bucket,
	}
}

func (f *LastBackupFinder) ListBlobs() (map[string]Blob, error) {
	directories, err := f.bucket.ListDirectories()
	if err != nil {
		return nil, fmt.Errorf("failed listing last backup blobs: %s", err)
	}

	blobsMap := map[string]Blob{}

	if len(directories) == 0 {
		return blobsMap, nil
	}

	isComplete, err := f.bucket.IsCompleteBackup(directories[0])
	if err != nil {
		return nil, fmt.Errorf("failed listing last backup blobs: %s", err)
	}

	if !isComplete {
		return blobsMap, nil
	}

	blobs, err := f.bucket.ListBlobs(directories[0])
	if err != nil {
		return nil, fmt.Errorf("failed listing last backup blobs: %s", err)
	}

	for _, blob := range blobs {
		nameParts := strings.Split(blob.Name, "/")
		blobsMap[nameParts[len(nameParts)-1]] = blob
	}

	return blobsMap, nil
}
