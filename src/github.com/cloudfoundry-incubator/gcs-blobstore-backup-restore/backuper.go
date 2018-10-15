package gcs

import (
	"fmt"
	"strings"
	"time"
)

type Backuper struct {
	buckets map[string]BucketPair
}

const liveBucketBackupArtifactName = "temporary-backup-artifact"

func NewBackuper(buckets map[string]BucketPair) Backuper {
	return Backuper{
		buckets: buckets,
	}
}

func (b *Backuper) CreateLiveBucketSnapshot() error {
	for _, bucketPair := range b.buckets {
		bucket := bucketPair.Bucket
		blobs, err := bucket.ListBlobs()

		if err != nil {
			return err
		}

		for _, blob := range blobs {
			_, err := bucket.CopyBlobWithinBucket(blob.Name, fmt.Sprintf("%s/%s", liveBucketBackupArtifactName, blob.Name))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (b *Backuper) TransferBlobsToBackupBucket() (map[string]BackupBucketAddress, error) {
	timestamp := time.Now().Format("2006_01_02_15_04_05")

	backupBuckets := make(map[string]BackupBucketAddress)

	for id, bucketPair := range b.buckets {
		bucket := bucketPair.Bucket
		backupBucket := bucketPair.BackupBucket

		backupBuckets[id] = BackupBucketAddress{
			BucketName: backupBucket.Name(),
			Path:       fmt.Sprintf("%s/%s/%s", backupBucket.Name(), timestamp, id),
		}

		backupBlobs, err := bucket.ListBlobs()

		if err != nil {
			return nil, err
		}

		var blobsToBeCleanedUp []Blob
		for _, blob := range backupBlobs {
			if strings.HasPrefix(blob.Name, fmt.Sprintf("%s/", liveBucketBackupArtifactName)) {
				blobsToBeCleanedUp = append(blobsToBeCleanedUp, blob)

				blobName := strings.Replace(blob.Name, liveBucketBackupArtifactName, fmt.Sprintf("%s/%s", timestamp, id), 1)

				_, err := bucket.CopyBlobBetweenBuckets(backupBucket, blob.Name, blobName)
				if err != nil {
					return nil, err
				}
			}
		}

		for _, blob := range blobsToBeCleanedUp {
			err = bucket.DeleteBlob(blob.Name)
			if err != nil {
				return nil, err
			}
		}

	}

	return backupBuckets, nil
}
