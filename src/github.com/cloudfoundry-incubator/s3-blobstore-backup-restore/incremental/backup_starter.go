package incremental

import (
	"fmt"
	"path/filepath"
)

type BucketPair struct {
	LiveBucket   Bucket
	BackupBucket Bucket
}

type BackupsToStart struct {
	BucketPair            BucketPair
	BackupDirectoryFinder BackupDirectoryFinder
}

//go:generate counterfeiter -o fakes/fake_clock.go . Clock
type Clock interface {
	Now() string
}

type BackupStarter struct {
	BackupsToStart map[string]BackupsToStart
	clock          Clock
}

func NewBackupStarter(backupsToStart map[string]BackupsToStart, clock Clock) BackupStarter {
	return BackupStarter{
		BackupsToStart: backupsToStart,
		clock:          clock,
	}
}

func (b BackupStarter) Run() error {
	timestamp := b.clock.Now()

	for bucketID, backupToStart := range b.BackupsToStart {
		backupDstPath := fmt.Sprintf("%s/%s", timestamp, bucketID)

		// find the last complete backup and list blobs
		backedUpBlobs, err := backupToStart.BackupDirectoryFinder.ListBlobs()
		if err != nil {
			return fmt.Errorf("failed to start backup: %s", err)
		}

		// list blobs in the live bucket
		liveBlobs, err := backupToStart.BucketPair.LiveBucket.ListBlobs(bucketID)
		if err != nil {
			return fmt.Errorf("failed to start backup: %s", err)
		}

		// copy new live blobs to the new backup directory
		err = backupToStart.BucketPair.copyNewLiveBlobsToBackup(backedUpBlobs, liveBlobs, backupDstPath)
		if err != nil {
			return fmt.Errorf("failed to copy blobs during backup: %s", err)
		}
	}

	// write the backup artifact for restore

	// write the backup directory and list of previously backed up blobs for backup completer

	return nil
}

func (b BucketPair) copyNewLiveBlobsToBackup(backedUpBlobs []BackedUpBlob, liveBlobs []Blob, backupDstPath string) error {
	backedUpBlobsMap := map[string]bool{}

	for _, backedUpBlob := range backedUpBlobs {
		backedUpBlobsMap[backedUpBlob.LiveBlobPath()] = true
	}

	for _, blob := range liveBlobs {
		_, exists := backedUpBlobsMap[blob.Path()]

		if !exists {
			path := filepath.Join(backupDstPath, blob.Path())
			err := b.LiveBucket.CopyBlobToBucket(b.BackupBucket, blob.Path(), path)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
