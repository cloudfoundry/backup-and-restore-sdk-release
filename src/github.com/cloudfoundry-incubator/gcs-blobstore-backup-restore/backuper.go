package gcs

import (
	"fmt"
	"strings"
	"time"
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

func (b *Backuper) Backup() (map[string]BucketBackup, error) {
	bucketBackups, previouslyBackedUpBlobs, err := b.CopyNewBlobs()
	if err != nil {
		return nil, err
	}

	err = b.CopyPreviouslyBackedUpBlobs(bucketBackups, previouslyBackedUpBlobs)
	if err != nil {
		return nil, err
	}

	return bucketBackups, nil
}

func (b *Backuper) CopyNewBlobs() (map[string]BucketBackup, map[string][]Blob, error) {
	timestamp := time.Now().Format(timestampFormat)
	bucketBackups := make(map[string]BucketBackup)
	previouslyBackedUpBlobs := make(map[string][]Blob)

	for bucketID, bucketPair := range b.bucketPairs {
		liveBucket := bucketPair.LiveBucket
		backupBucket := bucketPair.BackupBucket

		bucketBackups[bucketID] = BucketBackup{
			BucketName: backupBucket.Name(),
			Path:       fmt.Sprintf("%s/%s", timestamp, bucketID),
		}

		allBackupBlobs, err := bucketPair.BackupFinder.ListBlobs()
		if err != nil {
			return nil, nil, err
		}

		liveBlobs, err := liveBucket.ListBlobs("")
		if err != nil {
			return nil, nil, err
		}

		var previouslyBackedUpBucketBlobs []Blob
		for _, liveBlob := range liveBlobs {
			if previouslyBackedUpLiveBlob, ok := allBackupBlobs[liveBlob.Name]; ok {
				previouslyBackedUpBucketBlobs = append(previouslyBackedUpBucketBlobs, previouslyBackedUpLiveBlob)
			} else {
				err := liveBucket.CopyBlobToBucket(backupBucket, liveBlob.Name, fmt.Sprintf("%s/%s", bucketBackups[bucketID].Path, liveBlob.Name))
				if err != nil {
					return nil, nil, err
				}
			}
		}

		previouslyBackedUpBlobs[bucketID] = previouslyBackedUpBucketBlobs
	}
	return bucketBackups, previouslyBackedUpBlobs, nil
}

func (b *Backuper) CopyPreviouslyBackedUpBlobs(bucketBackups map[string]BucketBackup, previouslyBackedUpBlobs map[string][]Blob) error {
	for bucketID, bucketBackup := range bucketBackups {
		blobs, ok := previouslyBackedUpBlobs[bucketID]
		if !ok {
			return fmt.Errorf("cannot find previously backed up blobs for bucket id: %s", bucketID)
		}

		backupBucket := b.bucketPairs[bucketID].BackupBucket

		for _, blob := range blobs {
			nameParts := strings.Split(blob.Name, "/")
			destinationBlobName := fmt.Sprintf("%s/%s", bucketBackup.Path, nameParts[len(nameParts)-1])
			err := backupBucket.CopyBlobWithinBucket(blob.Name, destinationBlobName)
			if err != nil {
				return err
			}
		}

		err := backupBucket.MarkBackupComplete(bucketBackup.Path)
		if err != nil {
			return err
		}
	}

	return nil
}
