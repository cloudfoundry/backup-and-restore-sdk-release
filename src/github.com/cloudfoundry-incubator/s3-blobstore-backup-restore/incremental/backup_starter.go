package incremental

import "fmt"

type BucketPair struct {
	LiveBucket   Bucket
	BackupBucket Bucket
}

type BackupsToStart struct {
	BucketPair            BucketPair
	BackupDirectoryFinder BackupDirectoryFinder
}

type BackupStarter struct {
	BackupsToStart map[string]BackupsToStart
}

func (b BackupStarter) Run() error {
	for bucketID, backupToStart := range b.BackupsToStart {
		// find the last complete backup and list blobs
		_, err := backupToStart.BackupDirectoryFinder.ListBlobs(bucketID)
		if err != nil {
			return fmt.Errorf("failed to start backup: %s", err)
		}

		// list blobs in the live bucket

		// create a new backup directory

		// copy new live blobs to the new backup directory
	}

	// write the backup artifact for restore

	// write the backup directory and list of previously backed up blobs for backup completer

	return nil
}
