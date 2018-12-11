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

	completeDirectory, err := f.findLastCompleteDirectory(directories)
	if err != nil {
		return nil, fmt.Errorf("failed listing last backup blobs: %s", err)
	}

	if completeDirectory == "" {
		return map[string]Blob{}, nil
	}

	blobs, err := f.bucket.ListBlobs(completeDirectory)
	if err != nil {
		return nil, fmt.Errorf("failed listing last backup blobs: %s", err)
	}

	blobsMap := map[string]Blob{}
	for _, blob := range blobs {
		nameParts := strings.Split(blob.Name, "/")
		blobsMap[nameParts[len(nameParts)-1]] = blob
	}

	return blobsMap, nil
}

func (f *LastBackupFinder) findLastCompleteDirectory(directories []string) (string, error) {
	var completeDirectory string

	for i := len(directories) - 1; i >= 0; i-- {
		dir := directories[i]

		isComplete, err := f.bucket.IsBackupComplete(dir)
		if err != nil {
			return "", err
		}

		if isComplete {
			completeDirectory = dir
			break
		}
	}

	return completeDirectory, nil
}
