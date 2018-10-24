package gcs

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

//go:generate counterfeiter -o fakes/fake_backuper.go . Backuper
type Backuper interface {
	CreateLiveBucketSnapshot() (map[string]BackupBucketAddress, error)
	CopyBlobsWithinBackupBucket(backupBucketAddresses map[string]BackupBucketAddress) error
	TransferBlobsToBackupBucket() (map[string]BackupBucketAddress, error)
}

type GCSBackuper struct {
	buckets map[string]BucketPair
}

const liveBucketBackupArtifactName = "temporary-backup-artifact"
const commonBlobsName = "common_blobs.json"

func NewBackuper(buckets map[string]BucketPair) GCSBackuper {
	return GCSBackuper{
		buckets: buckets,
	}
}

func (b *GCSBackuper) CreateLiveBucketSnapshot() (map[string]BackupBucketAddress, error) {
	timestamp := time.Now().Format("2006_01_02_15_04_05")

	backupBuckets := make(map[string]BackupBucketAddress)

	for id, bucketPair := range b.buckets {
		var commonBlobs []Blob
		bucket := bucketPair.Bucket
		backupBucket := bucketPair.BackupBucket

		backupBuckets[id] = BackupBucketAddress{
			BucketName: backupBucket.Name(),
			Path:       fmt.Sprintf("%s/%s", timestamp, id),
		}

		blobs, err := bucket.ListBlobs()
		if err != nil {
			return nil, err
		}

		lastBackupBlobs, err := bucketPair.BackupBucket.ListLastBackupBlobs()
		if err != nil {
			return nil, err
		}

		inLastBackup := make(map[string]Blob)
		for _, blob := range lastBackupBlobs {
			nameParts := strings.Split(blob.Name, "/")
			inLastBackup[nameParts[len(nameParts)-1]] = blob
		}

		for _, blob := range blobs {
			if blobFromBackup, ok := inLastBackup[blob.Name]; ok {
				commonBlobs = append(commonBlobs, blobFromBackup)
			} else {
				err := bucket.CopyBlobBetweenBuckets(backupBucket, blob.Name, blob.Name)
				if err != nil {
					return nil, err
				}
			}
		}

		j, err := json.Marshal(commonBlobs)
		if err != nil {
			return nil, err
		}

		err = bucket.CreateFile(liveBucketBackupArtifactName+"/"+commonBlobsName, j)
		if err != nil {
			return nil, err
		}
	}
	return backupBuckets, nil
}

func (b *GCSBackuper) TransferBlobsToBackupBucket() (map[string]BackupBucketAddress, error) {
	timestamp := time.Now().Format("2006_01_02_15_04_05")

	backupBuckets := make(map[string]BackupBucketAddress)

	for id, bucketPair := range b.buckets {
		bucket := bucketPair.Bucket
		backupBucket := bucketPair.BackupBucket

		backupBlobs, err := bucket.ListBlobs()
		if err != nil {
			return nil, err
		}

		backupBuckets[id] = BackupBucketAddress{
			BucketName: backupBucket.Name(),
			Path:       fmt.Sprintf("%s/%s", timestamp, id),
		}
		for _, blob := range backupBlobs {
			if strings.HasPrefix(blob.Name, fmt.Sprintf("%s/", liveBucketBackupArtifactName)) {

				blobName := strings.Replace(blob.Name, liveBucketBackupArtifactName, fmt.Sprintf("%s/%s", timestamp, id), 1)

				err := bucket.CopyBlobBetweenBuckets(backupBucket, blob.Name, blobName)
				if err != nil {
					return nil, err
				}
			}
		}
	}

	return backupBuckets, nil
}

func (b *GCSBackuper) CopyBlobsWithinBackupBucket(backupBucketAddresses map[string]BackupBucketAddress) error {
	for bucketId, backupBucketAddress := range backupBucketAddresses {
		commonBlobBytes, err := b.buckets[bucketId].BackupBucket.GetBlob(backupBucketAddress.Path + "/common_blobs.json")
		if err != nil {
			return fmt.Errorf("failed to get %s/common_blobs.json: %v", backupBucketAddress.Path, err)
		}

		var commonBlobs []Blob
		err = json.Unmarshal(commonBlobBytes, &commonBlobs)
		if err != nil {
			return err
		}

		err = b.buckets[bucketId].BackupBucket.DeleteBlob(backupBucketAddress.Path + "/common_blobs.json")
		if err != nil {
			return err
		}

		for _, blob := range commonBlobs {
			nameParts := strings.Split(blob.Name, "/")
			destinationBlobName := fmt.Sprintf("%s/%s", backupBucketAddress.Path, nameParts[len(nameParts)-1])
			err = b.buckets[bucketId].BackupBucket.CopyBlobWithinBucket(blob.Name, destinationBlobName)
			if err != nil {
				return err
			}
		}

	}
	return nil
}
