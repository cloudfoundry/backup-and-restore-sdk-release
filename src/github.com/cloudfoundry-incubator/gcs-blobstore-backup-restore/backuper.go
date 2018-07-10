package gcs

import (
	"fmt"
)

type Backuper struct {
	buckets map[string]Bucket
}

func NewBackuper(buckets map[string]Bucket) Backuper {
	return Backuper{
		buckets: buckets,
	}
}

func (b *Backuper) Backup() (map[string]BucketBackup, error) {
	bucketBackups := map[string]BucketBackup{}

	for bucketIdentifier, bucket := range b.buckets {
		enabled, err := bucket.VersioningEnabled()
		if err != nil {
			return nil, err
		}

		if !enabled {
			return nil, fmt.Errorf("versioning is not enabled on bucket: %s", bucket.Name())
		}

		blobs, err := bucket.ListBlobs()
		if err != nil {
			return nil, err
		}

		bucketBackups[bucketIdentifier] = BucketBackup{
			Name:  bucket.Name(),
			Blobs: blobs,
		}
	}
	return bucketBackups, nil
}
