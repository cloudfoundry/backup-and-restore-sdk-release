package gcs

import (
	"fmt"
	"strings"
	"time"

	"github.com/cloudfoundry-incubator/bosh-backup-and-restore/executor"
)

const timestampFormat = "2006_01_02_15_04_05"

type Backuper struct {
	bucketPairs map[string]BucketPair
}

func NewBackuper(bucketPairs map[string]BucketPair) Backuper {
	return Backuper{
		bucketPairs: bucketPairs,
	}
}

type ExecutableBackup struct {
	blob         Blob
	backupAction func(Blob) error
}

func (e ExecutableBackup) Execute() error {
	return e.backupAction(e.blob)
}

func (b *Backuper) Backup() (map[string]BucketBackup, error) {
	timestamp := time.Now().Format(timestampFormat)
	bucketBackups := make(map[string]BucketBackup)

	for bucketID, bucketPair := range b.bucketPairs {
		liveBucket := bucketPair.LiveBucket
		backupBucket := bucketPair.BackupBucket

		bucketBackups[bucketID] = BucketBackup{
			BucketName: backupBucket.Name(),
			Path:       fmt.Sprintf("%s/%s", timestamp, bucketID),
		}

		liveBlobs, err := liveBucket.ListBlobs("")
		if err != nil {
			return nil, err
		}

		var executables []executor.Executable
		for _, liveBlob := range liveBlobs {
			executables = append(executables, ExecutableBackup{blob: liveBlob, backupAction: func(blob Blob) error {
				return liveBucket.CopyBlobToBucket(
					backupBucket,
					blob.Name(),
					fmt.Sprintf("%s/%s", bucketBackups[bucketID].Path, blob.Name()),
				)
			}})
		}

		errs := bucketPair.ExecutionStrategy.Run([][]executor.Executable{executables})
		if len(errs) != 0 {
			return map[string]BucketBackup{}, formatErrors(
				fmt.Sprintf("failed to backup bucket %s", liveBucket.Name()),
				errs,
			)
		}

	}

	return bucketBackups, nil
}

func formatErrors(contextString string, errors []error) error {
	errorStrings := make([]string, len(errors))
	for i, err := range errors {
		errorStrings[i] = err.Error()
	}
	return fmt.Errorf("%s: %s", contextString, strings.Join(errorStrings, "\n"))
}
