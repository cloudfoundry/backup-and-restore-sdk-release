package gcs

import (
	"fmt"
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

		for _, liveBlob := range liveBlobs {

			err := liveBucket.CopyBlobToBucket(backupBucket, liveBlob.Name(), fmt.Sprintf("%s/%s", bucketBackups[bucketID].Path, liveBlob.Name()))

			if err != nil {
				return nil, err
			}
		}

	}

	return bucketBackups, nil
}
