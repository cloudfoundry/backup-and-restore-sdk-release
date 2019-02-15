package incremental

import (
	"fmt"

	"github.com/cloudfoundry-incubator/bosh-backup-and-restore/executor"
)

type BucketPair struct {
	LiveBucket   Bucket
	BackupBucket Bucket
}

type BackupToStart struct {
	BucketPair            BucketPair
	BackupDirectoryFinder BackupDirectoryFinder
}

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
		backupDir := BackupDirectory{
			Path:   joinBlobPath(timestamp, bucketID),
			Bucket: backupToStart.BucketPair.BackupBucket,
		}

		backedUpBlobs, err := backupToStart.BackupDirectoryFinder.ListBlobs(bucketID)
		if err != nil {
			return fmt.Errorf("failed to start backup: %s", err)
		}

		liveBlobs, err := backupToStart.BucketPair.LiveBucket.ListBlobs("")
		if err != nil {
			return fmt.Errorf("failed to start backup: %s", err)
		}

		existingBlobsArtifact, err := backupToStart.BucketPair.copyNewLiveBlobsToBackup(backedUpBlobs, liveBlobs, backupDir.Path)
		if err != nil {
			return fmt.Errorf("failed to copy blobs during backup: %s", err)
		}

		backups[bucketID] = generateBackupArtifact(liveBlobs, backupDir)

		existingBlobs[bucketID] = generateExistingBlobsArtifact(
			existingBlobsArtifact,
			backupDir.Path,
		)
	}

	err := b.backupArtifact.Write(backups)
	if err != nil {
		return fmt.Errorf("failed to write backupArtifact: %s", err)
	}

	err = b.existingBlobsArtifact.Write(existingBlobs)
	if err != nil {
		return fmt.Errorf("failed to write existing blobs backupArtifact: %s", err)
	}

	return nil
}

func generateBackupArtifact(liveBlobs []Blob, dir BackupDirectory) Backup {
	var blobs []string
	for _, blob := range liveBlobs {
		blobs = append(blobs, joinBlobPath(dir.Path, blob.Path()))
	}

	return Backup{
		BucketName: dir.Bucket.Name(),
		Blobs:      blobs,
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
			Blobs: backedUpblobs,
			SrcBackupDirectoryPath: existingBlobs[0].BackupDirectoryPath,
			DstBackupDirectoryPath: dstBackupDirectoryPath,
		}
	}

	return Backup{
		DstBackupDirectoryPath: dstBackupDirectoryPath,
	}
}

func (b BucketPair) copyNewLiveBlobsToBackup(backedUpBlobs []BackedUpBlob, liveBlobs []Blob, backupDirPath string) ([]BackedUpBlob, error) {
	backedUpBlobsMap := make(map[string]BackedUpBlob)
	for _, backedUpBlob := range backedUpBlobs {
		backedUpBlobsMap[backedUpBlob.LiveBlobPath()] = backedUpBlob
	}

	var (
		executables   []executor.Executable
		existingBlobs []BackedUpBlob
	)
	for _, liveBlob := range liveBlobs {
		backedUpBlob, exists := backedUpBlobsMap[liveBlob.Path()]

		if !exists {
			executable := copyBlobFromBucketExecutable{
				src:       liveBlob.Path(),
				dst:       joinBlobPath(backupDirPath, liveBlob.Path()),
				srcBucket: b.LiveBucket,
				dstBucket: b.BackupBucket,
			}
			executables = append(executables, executable)
		} else {
			existingBlobs = append(existingBlobs, backedUpBlob)
		}
	}

	e := executor.NewParallelExecutor()
	e.SetMaxInFlight(200)

	errs := e.Run([][]executor.Executable{executables})
	if len(errs) != 0 {
		return nil, formatExecutorErrors("failing copying blobs in parallel", errs)
	}

	return existingBlobs, nil
}

type copyBlobFromBucketExecutable struct {
	src       string
	dst       string
	dstBucket Bucket
	srcBucket Bucket
}

func (e copyBlobFromBucketExecutable) Execute() error {
	return e.dstBucket.CopyBlobFromBucket(e.srcBucket, e.src, e.dst)
}
