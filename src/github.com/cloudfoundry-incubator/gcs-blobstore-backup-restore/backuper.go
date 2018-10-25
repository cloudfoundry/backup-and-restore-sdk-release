package gcs

import (
	"fmt"
	"strings"
	"time"
)

type Backuper struct {
	buckets map[string]BucketPair
}

func NewBackuper(buckets map[string]BucketPair) Backuper {
	return Backuper{
		buckets: buckets,
	}
}

func (b *Backuper) CreateLiveBucketSnapshot() (map[string]BackupBucketDir, map[string][]Blob, error) {
	timestamp := time.Now().Format("2006_01_02_15_04_05")
	backupBuckets := make(map[string]BackupBucketDir)
	allCommonBlobs := make(map[string][]Blob)

	for bucketId, bucketPair := range b.buckets {
		var bucketCommonBlobs []Blob
		bucket := bucketPair.Bucket
		backupBucket := bucketPair.BackupBucket

		backupBuckets[bucketId] = BackupBucketDir{
			BucketName: backupBucket.Name(),
			Path:       fmt.Sprintf("%s/%s", timestamp, bucketId),
		}

		lastBackupBlobs, err := backupBucket.LastBackupBlobs()
		if err != nil {
			return nil, nil, err
		}

		blobs, err := bucket.ListBlobs()
		if err != nil {
			return nil, nil, err
		}

		for _, blob := range blobs {
			if blobFromBackup, ok := lastBackupBlobs[blob.Name]; ok {
				bucketCommonBlobs = append(bucketCommonBlobs, blobFromBackup)
			} else {
				err := bucket.CopyBlobBetweenBuckets(backupBucket, blob.Name, fmt.Sprintf("%s/%s", backupBuckets[bucketId].Path, blob.Name))
				if err != nil {
					return nil, nil, err
				}
			}
		}

		allCommonBlobs[bucketId] = bucketCommonBlobs
	}
	return backupBuckets, allCommonBlobs, nil
}

func (b *Backuper) CopyBlobsWithinBackupBucket(backupBucketAddresses map[string]BackupBucketDir, commonBlobs map[string][]Blob) error {
	for bucketId, backupBucketAddress := range backupBucketAddresses {
		commonBlobList, ok := commonBlobs[bucketId]
		if !ok {
			return fmt.Errorf("cannot find commonBlobs for bucket id: %s", bucketId)
		}

		for _, blob := range commonBlobList {
			nameParts := strings.Split(blob.Name, "/")
			destinationBlobName := fmt.Sprintf("%s/%s", backupBucketAddress.Path, nameParts[len(nameParts)-1])
			err := b.buckets[bucketId].BackupBucket.CopyBlobWithinBucket(blob.Name, destinationBlobName)
			if err != nil {
				return err
			}
		}

	}
	return nil
}
