package gcs

import (
	"fmt"
	"strings"
	"time"
)

type Backuper struct {
	bucketPairs map[string]BucketPair
}

func NewBackuper(bucketPairs map[string]BucketPair) Backuper {
	return Backuper{
		bucketPairs: bucketPairs,
	}
}

func (b *Backuper) Backup() (map[string]BackupBucketDirectory, error) {
	backupBucketDirectories, commonBlobs, err := b.CreateLiveBucketSnapshot()
	if err != nil {
		return nil, err
	}

	err = b.CopyBlobsWithinBackupBucket(backupBucketDirectories, commonBlobs)
	if err != nil {
		return nil, err
	}

	return backupBucketDirectories, nil
}

func (b *Backuper) CreateLiveBucketSnapshot() (map[string]BackupBucketDirectory, map[string][]Blob, error) {
	timestamp := time.Now().Format("2006_01_02_15_04_05")
	backupBucketDirectories := make(map[string]BackupBucketDirectory)
	allCommonBlobs := make(map[string][]Blob)

	for bucketId, bucketPair := range b.bucketPairs {
		var bucketCommonBlobs []Blob
		liveBucket := bucketPair.LiveBucket
		backupBucket := bucketPair.BackupBucket

		backupBucketDirectories[bucketId] = BackupBucketDirectory{
			BucketName: backupBucket.Name(),
			Path:       fmt.Sprintf("%s/%s", timestamp, bucketId),
		}

		lastBackupBlobs, err := bucketPair.BackupFinder.ListBlobs()
		if err != nil {
			return nil, nil, err
		}

		blobs, err := liveBucket.ListBlobs("")
		if err != nil {
			return nil, nil, err
		}

		for _, blob := range blobs {
			if blobFromBackup, ok := lastBackupBlobs[blob.Name]; ok {
				bucketCommonBlobs = append(bucketCommonBlobs, blobFromBackup)
			} else {
				err := liveBucket.CopyBlobBetweenBuckets(backupBucket, blob.Name, fmt.Sprintf("%s/%s", backupBucketDirectories[bucketId].Path, blob.Name))
				if err != nil {
					return nil, nil, err
				}
			}
		}

		allCommonBlobs[bucketId] = bucketCommonBlobs
	}
	return backupBucketDirectories, allCommonBlobs, nil
}

func (b *Backuper) CopyBlobsWithinBackupBucket(backupBucketAddresses map[string]BackupBucketDirectory, commonBlobs map[string][]Blob) error {
	for bucketId, backupBucketAddress := range backupBucketAddresses {
		commonBlobList, ok := commonBlobs[bucketId]
		if !ok {
			return fmt.Errorf("cannot find commonBlobs for bucket id: %s", bucketId)
		}

		backupBucket := b.bucketPairs[bucketId].BackupBucket

		for _, blob := range commonBlobList {
			nameParts := strings.Split(blob.Name, "/")
			destinationBlobName := fmt.Sprintf("%s/%s", backupBucketAddress.Path, nameParts[len(nameParts)-1])
			err := backupBucket.CopyBlobWithinBucket(blob.Name, destinationBlobName)
			if err != nil {
				return err
			}
		}

		err := backupBucket.CreateBackupCompleteBlob(backupBucketAddress.Path)
		if err != nil {
			return err
		}
	}

	return nil
}
