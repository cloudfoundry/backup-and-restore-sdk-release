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

func (r Restorer) Restore(bucketBackups map[string]BucketBackup) error {
	for bucketID := range bucketBackups {
		_, ok := r.buckets[bucketID]
		if !ok {
			return fmt.Errorf("no entry found in restore config for bucket: %s", bucketID)
		}
	}

	for bucketID := range r.buckets {
		_, ok := bucketBackups[bucketID]
		if !ok {
			return fmt.Errorf("no entry found in restore artifact for bucket: %s", bucketID)
		}
	}

	for bucketID, bucketBackup := range bucketBackups {
		err := r.buckets[bucketID].BackupBucket.CopyBlobsToBucket(
			r.buckets[bucketID].LiveBucket,
			bucketBackup.Path,
		)
		if err != nil {
			return err
		}
	}

	return nil
}
