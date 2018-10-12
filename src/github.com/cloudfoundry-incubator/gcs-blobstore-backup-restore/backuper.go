package gcs

import (
	"fmt"
	"strings"
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

func (b *Backuper) TransferBlobsToBackupBucket() error {
	for _, bucketPair := range b.buckets {
		bucket := bucketPair.Bucket
		backupBucket := bucketPair.BackupBucket

		backupBlobs, err := bucket.ListBlobs()

		if err != nil {
			return err
		}

		var blobsToBeCleanedUp []Blob
		for _, blob := range backupBlobs {
			if strings.HasPrefix(blob.Name, fmt.Sprintf("%s/", liveBucketBackupArtifactName)) {
				blobsToBeCleanedUp = append(blobsToBeCleanedUp, blob)
				_, err := bucket.CopyBlobBetweenBuckets(backupBucket, blob.Name, blob.Name)
				if err != nil {
					return err
				}
			}
		}

		for _, blob := range blobsToBeCleanedUp {
			err = bucket.Delete(blob.Name)
			if err != nil {
				return err
			}
		}

	}

	return nil
}
