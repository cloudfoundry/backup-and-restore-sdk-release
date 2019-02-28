package incremental

import (
	"fmt"
)

//go:generate counterfeiter -o fakes/fake_clock.go . Clock
type Clock interface {
	Now() string
}

type BackupStarter struct {
	BackupsToStart        map[string]BackupToStart
	clock                 Clock
	backupArtifact        Artifact
	existingBlobsArtifact Artifact
}

func NewBackupStarter(backupsToStart map[string]BackupToStart, clock Clock, backupArtifact, existingBlobsArtifact Artifact) BackupStarter {
	return BackupStarter{
		BackupsToStart:        backupsToStart,
		clock:                 clock,
		backupArtifact:        backupArtifact,
		existingBlobsArtifact: existingBlobsArtifact,
	}
}

func (b BackupStarter) Run() error {
	timestamp := b.clock.Now()

	backups := map[string]Backup{}
	existingBlobs := map[string]Backup{}

	for bucketID, backupToStart := range b.BackupsToStart {
		if backupToStart.SameAsBucketID != "" {
			sameAsBackup := Backup{SameBucketAs: backupToStart.SameAsBucketID}
			backups[bucketID] = sameAsBackup
			existingBlobs[bucketID] = sameAsBackup
			continue
		}

		backupDir := BackupDirectory{
			Path:   joinBlobPath(timestamp, bucketID),
			Bucket: backupToStart.BucketPair.ConfigBackupBucket,
		}

		backedUpBlobs, err := backupToStart.BackupDirectoryFinder.ListBlobs(bucketID, backupToStart.BucketPair.ConfigBackupBucket)
		if err != nil {
			return fmt.Errorf("failed to start backup: %s", err)
		}

		liveBlobs, err := backupToStart.BucketPair.ConfigLiveBucket.ListBlobs("")
		if err != nil {
			return fmt.Errorf("failed to start backup: %s", err)
		}

		filteredLiveBlobs := filterOutBackupComplete(liveBlobs)

		existingBlobsArtifact, err := backupToStart.BucketPair.CopyNewLiveBlobsToBackup(backedUpBlobs, filteredLiveBlobs, backupDir.Path)
		if err != nil {
			return fmt.Errorf("failed to copy blobs during backup: %s", err)
		}

		backups[bucketID] = generateBackupArtifact(filteredLiveBlobs, backupDir)

		existingBlobs[bucketID] = generateExistingBlobsArtifact(
			existingBlobsArtifact,
			backupDir.Path,
		)
	}

	err := b.backupArtifact.Write(backups)
	if err != nil {
		return fmt.Errorf("failed to write backup artifact: %s", err)
	}

	err = b.existingBlobsArtifact.Write(existingBlobs)
	if err != nil {
		return fmt.Errorf("failed to write existing blobs artifact: %s", err)
	}

	return nil
}

func filterOutBackupComplete(liveBlobs []Blob) []Blob {
	var filteredLiveBlobs []Blob

	for _, liveBlob := range liveBlobs {
		if liveBlob.Path() != backupComplete {
			filteredLiveBlobs = append(filteredLiveBlobs, liveBlob)
		}
	}

	return filteredLiveBlobs
}

func generateBackupArtifact(liveBlobs []Blob, dir BackupDirectory) Backup {
	var blobs []string
	for _, blob := range liveBlobs {
		blobs = append(blobs, joinBlobPath(dir.Path, blob.Path()))
	}

	return Backup{
		BucketName:             dir.Bucket.Name(),
		Blobs:                  blobs,
		SrcBackupDirectoryPath: dir.Path,
		BucketRegion:           dir.Bucket.Region(),
	}
}

func generateExistingBlobsArtifact(existingBlobs []BackedUpBlob, dstBackupDirectoryPath string) Backup {
	if len(existingBlobs) != 0 {
		var backedUpblobs []string
		for _, blob := range existingBlobs {
			backedUpblobs = append(backedUpblobs, blob.Path)
		}

		return Backup{
			Blobs:                  backedUpblobs,
			SrcBackupDirectoryPath: existingBlobs[0].BackupDirectoryPath,
			DstBackupDirectoryPath: dstBackupDirectoryPath,
		}
	}

	return Backup{
		DstBackupDirectoryPath: dstBackupDirectoryPath,
	}
}
