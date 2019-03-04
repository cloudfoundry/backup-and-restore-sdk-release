package gcs

import "fmt"

type Restorer struct {
	backupsToComplete map[string]BackupToComplete
}

func NewRestorer(backupsToComplete map[string]BackupToComplete) Restorer {
	return Restorer{
		backupsToComplete: backupsToComplete,
	}
}

func (r Restorer) Restore(bucketBackups map[string]BucketBackup) error {
	for bucketID := range bucketBackups {
		_, ok := r.backupsToComplete[bucketID]
		if !ok {
			return fmt.Errorf("no entry found in restore config for bucket: %s", bucketID)
		}
	}

	for bucketID := range r.backupsToComplete {
		_, ok := bucketBackups[bucketID]
		if !ok {
			return fmt.Errorf("no entry found in restore artifact for bucket: %s", bucketID)
		}
	}

	for bucketID, bucketBackup := range bucketBackups {
		err := r.backupsToComplete[bucketID].BucketPair.BackupBucket.CopyBlobsToBucket(
			r.backupsToComplete[bucketID].BucketPair.LiveBucket,
			bucketBackup.Path,
		)
		if err != nil {
			return err
		}
	}

	return nil
}
