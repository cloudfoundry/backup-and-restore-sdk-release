package incremental

import (
	"fmt"
	"strings"
)

const blobDelimiter = "/"

type BucketPair struct {
	LiveBucket   Bucket
	BackupBucket Bucket
}

//go:generate counterfeiter -o fakes/fake_bucket.go . Bucket
type Bucket interface {
	Name() string
	ListBlobs() ([]Blob, error)
	ListDirectories() ([]string, error)
	CopyBlobWithinBucket(src, dst string) error
	CopyBlobToBucket(bucket Bucket, src, dst string) error
}

//go:generate counterfeiter -o fakes/fake_blob.go . Blob
type Blob interface {
	Name() string
}

// BackupDirectory is serializable
// requires fields to make a bucket: name, region, creds etc.
//go:generate counterfeiter -o fakes/fake_backup_directory.go . BackupDirectory
type BackupDirectory interface {
	Path() string
	IsComplete() (bool, error)
	MarkComplete() error
}

type BackupToComplete struct {
	BackupBucket    Bucket
	BackupDirectory BackupDirectory
	BlobsToCopy     []Blob
}

type BackupCompleter struct {
	BackupsToComplete map[string]BackupToComplete
}

func (b BackupCompleter) Run() error {
	for bucketID, backupToComplete := range b.BackupsToComplete {
		for _, blob := range backupToComplete.BlobsToCopy {
			parts := strings.SplitN(blob.Name(), bucketID+blobDelimiter, 2)
			dst := strings.Join([]string{backupToComplete.BackupDirectory.Path(), bucketID, parts[1]}, blobDelimiter)

			err := backupToComplete.BackupBucket.CopyBlobWithinBucket(blob.Name(), dst)
			if err != nil {

				return fmt.Errorf("failed to complete backup: %s", err)
			}
		}

		err := backupToComplete.BackupDirectory.MarkComplete()
		if err != nil {
			return fmt.Errorf("failed to complete backup: %s", err)
		}
	}

	return nil
}
