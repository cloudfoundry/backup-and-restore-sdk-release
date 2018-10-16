package gcs

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Backuper struct {
	buckets map[string]BucketPair
}

const liveBucketBackupArtifactName = "temporary-backup-artifact"
const commonBlobsName = "common_blobs.json"

func NewBackuper(buckets map[string]BucketPair) Backuper {
	return Backuper{
		buckets: buckets,
	}
}

func (b *Backuper) CreateLiveBucketSnapshot() error {
	for _, bucketPair := range b.buckets {
		var commonBlobs []Blob
		bucket := bucketPair.Bucket

		blobs, err := bucket.ListBlobs()
		if err != nil {
			return err
		}

		lastBackupBlobs, err := bucketPair.BackupBucket.ListLastBackupBlobs()
		if err != nil {
			return err
		}

		inLastBackup := make(map[string]bool)
		for _, blob := range lastBackupBlobs {
			nameParts := strings.Split(blob.Name, "/")
			inLastBackup[nameParts[len(nameParts)-1]] = true
		}

		for _, blob := range blobs {
			if inLastBackup[blob.Name] {
				commonBlobs = append(commonBlobs, blob)
			} else {
				_, err := bucket.CopyBlobWithinBucket(blob.Name, fmt.Sprintf("%s/%s", liveBucketBackupArtifactName, blob.Name))
				if err != nil {
					return err
				}
			}
		}

		j, err := json.Marshal(commonBlobs)
		if err != nil {
			return err
		}

		_, err = bucket.CreateFile(liveBucketBackupArtifactName+"/"+commonBlobsName, j)
		if err != nil {
			return err
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

func (b *Backuper) CopyBlobsWithinBackupBucket() error {
	return nil
}
