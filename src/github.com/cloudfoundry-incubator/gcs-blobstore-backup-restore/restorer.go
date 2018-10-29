package gcs

import "fmt"

type Restorer struct {
	buckets map[string]BucketPair
}

func NewRestorer(buckets map[string]BucketPair) Restorer {
	return Restorer{
		buckets: buckets,
	}
}

func (r Restorer) Restore(backupArtifact map[string]BackupBucketDirectory) error {
	for bucketID := range backupArtifact {
		_, ok := r.buckets[bucketID]
		if !ok {
			return fmt.Errorf("no entry found in restore config for bucket: %s", bucketID)
		}
	}

	for bucketID := range r.buckets {
		_, ok := backupArtifact[bucketID]
		if !ok {
			return fmt.Errorf("no entry found in restore artifact for bucket: %s", bucketID)
		}
	}

	for bucketID, backupBucketDirectory := range backupArtifact {
		err := r.buckets[bucketID].BackupBucket.CopyBlobsBetweenBuckets(
			r.buckets[bucketID].Bucket,
			backupBucketDirectory.Path,
		)
		if err != nil {
			return err
		}
	}

	return nil
}
