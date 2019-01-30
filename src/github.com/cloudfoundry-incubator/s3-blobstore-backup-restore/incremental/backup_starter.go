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
	BackupsToStart         map[string]BackupsToStart
	clock                  Clock
	backupArtifact         Artifact
	reuseableBlobsArtifact Artifact
}

func NewBackupStarter(backupsToStart map[string]BackupsToStart, clock Clock, backupArtifact, reuseableBlobsArtifact Artifact) BackupStarter {
	return BackupStarter{
		BackupsToStart:         backupsToStart,
		clock:                  clock,
		backupArtifact:         backupArtifact,
		reuseableBlobsArtifact: reuseableBlobsArtifact,
	}
}

func (b BackupStarter) Run() error {
	timestamp := b.clock.Now()

	bucketBackups := map[string]BucketBackup{}
	reusableBucketBackups := map[string]BucketBackup{}

	for bucketID, backupToStart := range b.BackupsToStart {
		backupDstPath := fmt.Sprintf("%s/%s", timestamp, bucketID)

		backedUpBlobs, err := backupToStart.BackupDirectoryFinder.ListBlobs()
		if err != nil {
			return fmt.Errorf("failed to start backup: %s", err)
		}

		liveBlobs, err := backupToStart.BucketPair.LiveBucket.ListBlobs(bucketID)
		if err != nil {
			return fmt.Errorf("failed to start backup: %s", err)
		}

		reuseableBlobsArtifact, err := backupToStart.BucketPair.copyNewLiveBlobsToBackup(backedUpBlobs, liveBlobs, backupDstPath)
		if err != nil {
			return fmt.Errorf("failed to copy blobs during backup: %s", err)
		}

		bucketBackups[bucketID] = generateBackupArtifact(
			liveBlobs,
			backupDstPath,
			backupToStart.BucketPair.BackupBucket.Name(),
		)

		reusableBucketBackups[bucketID] = generateReuseableBlobsArtifact(
			reuseableBlobsArtifact,
			backupToStart.BucketPair.BackupBucket.Name(),
		)
	}

	err := b.backupArtifact.Write(bucketBackups)
	if err != nil {
		return fmt.Errorf("failed to write backupArtifact: %s", err)
	}

	err = b.reuseableBlobsArtifact.Write(reusableBucketBackups)
	if err != nil {
		return fmt.Errorf("failed to write existing blobs backupArtifact: %s", err)
	}

	return nil
}

func generateBackupArtifact(liveBlobs []Blob, backupDstPath string, bucketName string) BucketBackup {
	blobs := []string{}
	for _, blob := range liveBlobs {
		blobs = append(blobs, filepath.Join(backupDstPath, blob.Path()))
	}
	return BucketBackup{
		BucketName:          bucketName,
		Blobs:               blobs,
		BackupDirectoryPath: backupDstPath,
	}
}

func generateReuseableBlobsArtifact(reuseableBlobs []BackedUpBlob, bucketName string) BucketBackup {
	if len(reuseableBlobs) != 0 {
		backedUpblobs := []string{}
		for _, blob := range reuseableBlobs {
			backedUpblobs = append(backedUpblobs, blob.Path)
		}

		return BucketBackup{
			BucketName:          bucketName,
			Blobs:               backedUpblobs,
			BackupDirectoryPath: reuseableBlobs[0].BackupDirectoryPath,
		}
	}

	return BucketBackup{}
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
