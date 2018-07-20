package gcs

import (
	"fmt"
)

type Restorer struct {
	buckets map[string]Bucket
}

func NewRestorer(buckets map[string]Bucket) Restorer {
	return Restorer{buckets: buckets}
}

func (r Restorer) Restore(backups map[string]BucketBackup) error {
	for bucketIdentifier := range backups {
		_, exists := r.buckets[bucketIdentifier]
		if !exists {
			return fmt.Errorf("bucket identifier '%s' not found in buckets configuration", bucketIdentifier)
		}
	}

	for _, bucket := range r.buckets {
		enabled, err := bucket.VersioningEnabled()
		if err != nil {
			return fmt.Errorf("failed to check if versioning is enabled on bucket '%s': %s", bucket.Name(), err)
		}

		if !enabled {
			return fmt.Errorf("versioning is not enabled on bucket '%s'", bucket.Name())
		}
	}

	for bucketIdentifier, backup := range backups {
		bucket := r.buckets[bucketIdentifier]

		for _, blob := range backup.Blobs {
			err := bucket.CopyVersion(blob, backup.Name)
			if err != nil {
				return fmt.Errorf("failed to copy blob '%s': %s", blob.Name, err)
			}
		}
	}

	return nil
}
