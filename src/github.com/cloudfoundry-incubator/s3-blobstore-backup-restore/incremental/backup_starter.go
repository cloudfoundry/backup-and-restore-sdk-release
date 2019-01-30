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
	BackupsToStart        map[string]BackupsToStart
	clock                 Clock
	artifact              Artifact
	existingBlobsArtifact Artifact
}

func NewBackupStarter(backupsToStart map[string]BackupsToStart, clock Clock, artifact, existingBlobsArtifact Artifact) BackupStarter {
	return BackupStarter{
		BackupsToStart:        backupsToStart,
		clock:                 clock,
		artifact:              artifact,
		existingBlobsArtifact: existingBlobsArtifact,
	}
}

func (b BackupStarter) Run() error {
	timestamp := b.clock.Now()

	bucketBackups := map[string]BucketBackup{}
	existingBucketBackups := map[string]BucketBackup{}

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
		existingBlobs, err := backupToStart.BucketPair.copyNewLiveBlobsToBackup(backedUpBlobs, liveBlobs, backupDstPath)
		if err != nil {
			return fmt.Errorf("failed to copy blobs during backup: %s", err)
		}

		blobs := []string{}
		for _, blob := range liveBlobs {
			blobs = append(blobs, filepath.Join(backupDstPath, blob.Path()))
		}

		bucketBackups[bucketID] = BucketBackup{
			BucketName:          backupToStart.BucketPair.BackupBucket.Name(),
			Blobs:               blobs,
			BackupDirectoryPath: backupDstPath,
		}

		if len(existingBlobs) != 0 {
			backedUpblobs := []string{}
			for _, blob := range existingBlobs {
				backedUpblobs = append(backedUpblobs, blob.Path)
			}

			existingBucketBackups[bucketID] = BucketBackup{
				BucketName:          backupToStart.BucketPair.BackupBucket.Name(),
				Blobs:               backedUpblobs,
				BackupDirectoryPath: existingBlobs[0].BackupDirectoryPath,
			}
		}
	}

	// write the backup artifact for restore
	err := b.artifact.Write(bucketBackups)
	if err != nil {
		return fmt.Errorf("failed to write artifact: %s", err)
	}

	// write the backup directory and list of previously backed up blobs for backup completer
	err = b.existingBlobsArtifact.Write(existingBucketBackups)
	if err != nil {
		return fmt.Errorf("failed to write existing blobs artifact: %s", err)
	}

	return nil
}

func (b BucketPair) copyNewLiveBlobsToBackup(backedUpBlobs []BackedUpBlob, liveBlobs []Blob, backupDstPath string) ([]BackedUpBlob, error) {
	existingBlobs := []BackedUpBlob{}
	backedUpBlobsMap := map[string]BackedUpBlob{}

	for _, backedUpBlob := range backedUpBlobs {
		backedUpBlobsMap[backedUpBlob.LiveBlobPath()] = backedUpBlob
	}

	for _, blob := range liveBlobs {
		backedUpBlob, exists := backedUpBlobsMap[blob.Path()]

		if !exists {
			path := filepath.Join(backupDstPath, blob.Path())
			err := b.LiveBucket.CopyBlobToBucket(b.BackupBucket, blob.Path(), path)
			if err != nil {
				return nil, err
			}
		} else {
			existingBlobs = append(existingBlobs, backedUpBlob)
		}

	}

	return existingBlobs, nil
}
