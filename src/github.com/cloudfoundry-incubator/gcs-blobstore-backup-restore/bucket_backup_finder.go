package gcs

import (
	"fmt"
	"strings"
)

//go:generate counterfeiter -o fakes/fake_bucket_backup_finder.go . BucketBackupFinder
type BucketBackupFinder interface {
	ListBlobs() (map[string]Blob, error)
}

type LastBucketBackupFinder struct {
	bucket   Bucket
	bucketID string
}

func NewLastBucketBackupFinder(bucketID string, bucket Bucket) *LastBucketBackupFinder {
	return &LastBucketBackupFinder{
		bucket:   bucket,
		bucketID: bucketID,
	}
}

func (f *LastBucketBackupFinder) ListBlobs() (map[string]Blob, error) {
	backups, err := f.bucket.ListBackups()
	if err != nil {
		return nil, fmt.Errorf("failed listing last backup blobs: %s", err)
	}

	completeBucketBackup, err := f.findLastComplete(backups)
	if err != nil {
		return nil, fmt.Errorf("failed listing last backup blobs: %s", err)
	}

	if completeBucketBackup == "" {
		return map[string]Blob{}, nil
	}

	blobs, err := f.bucket.ListBlobs(completeBucketBackup)
	if err != nil {
		return nil, fmt.Errorf("failed listing last backup blobs: %s", err)
	}

	blobsMap := map[string]Blob{}
	for _, blob := range blobs {
		nameParts := strings.Split(blob.Name(), "/")
		blobsMap[nameParts[len(nameParts)-1]] = blob
	}

	return blobsMap, nil
}

func (f *LastBucketBackupFinder) findLastComplete(backups []string) (string, error) {
	var completeBucketBackup string

	for i := len(backups) - 1; i >= 0; i-- {
		dir := backups[i]

		isComplete, err := f.bucket.IsBackupComplete(fmt.Sprintf("%s/%s", dir, f.bucketID))
		if err != nil {
			return "", err
		}

		if isComplete {
			completeBucketBackup = dir
			break
		}
	}

	return completeBucketBackup, nil
}
