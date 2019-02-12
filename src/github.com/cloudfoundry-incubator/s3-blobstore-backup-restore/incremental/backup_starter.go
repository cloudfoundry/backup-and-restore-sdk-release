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
	BackupsToStart         map[string]BackupToStart
	clock                  Clock
	backupArtifact         Artifact
	reuseableBlobsArtifact Artifact
}

func NewBackupStarter(backupsToStart map[string]BackupToStart, clock Clock, backupArtifact, reuseableBlobsArtifact Artifact) BackupStarter {
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

		backedUpBlobs, err := backupToStart.BackupDirectoryFinder.ListBlobs(bucketID)
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
