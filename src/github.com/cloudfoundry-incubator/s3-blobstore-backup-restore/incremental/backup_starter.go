package incremental

import (
	"fmt"
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
		backupDir := BackupDirectory{
			Path:   joinBlobPath(timestamp, bucketID),
			Bucket: backupToStart.BucketPair.BackupBucket,
		}

		backedUpBlobs, err := backupToStart.BackupDirectoryFinder.ListBlobs()
		if err != nil {
			return fmt.Errorf("failed to start backup: %s", err)
		}

		liveBlobs, err := backupToStart.BucketPair.LiveBucket.ListBlobs("")
		if err != nil {
			return fmt.Errorf("failed to start backup: %s", err)
		}

		reuseableBlobsArtifact, err := backupToStart.BucketPair.copyNewLiveBlobsToBackup(backedUpBlobs, liveBlobs, backupDir.Path)
		if err != nil {
			return fmt.Errorf("failed to copy blobs during backup: %s", err)
		}

		bucketBackups[bucketID] = generateBackupArtifact(liveBlobs, backupDir)

		reusableBucketBackups[bucketID] = generateReuseableBlobsArtifact(
			reuseableBlobsArtifact,
			backupToStart.BucketPair.BackupBucket.Name(),
			backupToStart.BucketPair.BackupBucket.Region(),
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

func generateBackupArtifact(liveBlobs []Blob, dir BackupDirectory) BucketBackup {
	var blobs []string
	for _, blob := range liveBlobs {
		blobs = append(blobs, joinBlobPath(dir.Path, blob.Path()))
	}

	return BucketBackup{
		BucketName:          dir.Bucket.Name(),
		Blobs:               blobs,
		BackupDirectoryPath: dir.Path,
		BucketRegion:        dir.Bucket.Region(),
	}
}

func generateReuseableBlobsArtifact(reuseableBlobs []BackedUpBlob, bucketName, region string) BucketBackup {
	if len(reuseableBlobs) != 0 {
		var backedUpblobs []string
		for _, blob := range reuseableBlobs {
			backedUpblobs = append(backedUpblobs, blob.Path)
		}

		return BucketBackup{
			BucketName:          bucketName,
			BucketRegion:        region,
			Blobs:               backedUpblobs,
			BackupDirectoryPath: reuseableBlobs[0].BackupDirectoryPath,
		}
	}

	return BucketBackup{}
}

func (b BucketPair) copyNewLiveBlobsToBackup(backedUpBlobs []BackedUpBlob, liveBlobs []Blob, backupDstPath string) ([]BackedUpBlob, error) {
	var existingBlobs []BackedUpBlob
	backedUpBlobsMap := map[string]BackedUpBlob{}

	for _, backedUpBlob := range backedUpBlobs {
		backedUpBlobsMap[backedUpBlob.LiveBlobPath()] = backedUpBlob
	}

	for _, liveBlob := range liveBlobs {
		backedUpBlob, exists := backedUpBlobsMap[liveBlob.Path()]

		if !exists {
			dstBlobPath := joinBlobPath(backupDstPath, liveBlob.Path())
			err := b.BackupBucket.CopyBlobFromBucket(b.LiveBucket, liveBlob.Path(), dstBlobPath)
			if err != nil {
				return nil, err
			}
		} else {
			existingBlobs = append(existingBlobs, backedUpBlob)
		}
	}

	return existingBlobs, nil
}
