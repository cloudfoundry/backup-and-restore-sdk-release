package gcs

import (
	"fmt"
	"strings"
	"time"
)

//go:generate counterfeiter -o fakes/fake_backuper.go . Backuper
type Backuper interface {
	CreateLiveBucketSnapshot() (map[string]BackupBucketAddress, map[string][]Blob, error)
	CopyBlobsWithinBackupBucket(map[string]BackupBucketAddress, map[string][]Blob) error
	TransferBlobsToBackupBucket() (map[string]BackupBucketAddress, error)
}

type GCSBackuper struct {
	buckets map[string]BucketPair
}

const liveBucketBackupArtifactName = "temporary-backup-artifact"

func NewBackuper(buckets map[string]BucketPair) GCSBackuper {
	return GCSBackuper{
		buckets: buckets,
	}
}

func (b *GCSBackuper) CreateLiveBucketSnapshot() (map[string]BackupBucketAddress, map[string][]Blob, error) {
	timestamp := time.Now().Format("2006_01_02_15_04_05")

	backupBuckets := make(map[string]BackupBucketAddress)

	commonBlobs := make(map[string][]Blob)

	for id, bucketPair := range b.buckets {
		var commonBlobList []Blob
		bucket := bucketPair.Bucket
		backupBucket := bucketPair.BackupBucket

		backupBuckets[id] = BackupBucketAddress{
			BucketName: backupBucket.Name(),
			Path:       fmt.Sprintf("%s/%s", timestamp, id),
		}

		blobs, err := bucket.ListBlobs()
		if err != nil {
			return nil, nil, err
		}

		lastBackupBlobs, err := bucketPair.BackupBucket.ListLastBackupBlobs()
		if err != nil {
			return nil, nil, err
		}

		inLastBackup := make(map[string]Blob)
		for _, blob := range lastBackupBlobs {
			nameParts := strings.Split(blob.Name, "/")
			inLastBackup[nameParts[len(nameParts)-1]] = blob
		}

		for _, blob := range blobs {
			if blobFromBackup, ok := inLastBackup[blob.Name]; ok {
				commonBlobList = append(commonBlobList, blobFromBackup)
			} else {
				err := bucket.CopyBlobBetweenBuckets(backupBucket, blob.Name, fmt.Sprintf("%s/%s", backupBuckets[id].Path, blob.Name))
				if err != nil {
					return nil, nil, err
				}
			}
		}

		commonBlobs[id] = commonBlobList
	}
	return backupBuckets, commonBlobs, nil
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

func (b *GCSBackuper) CopyBlobsWithinBackupBucket(backupBucketAddresses map[string]BackupBucketAddress, commonBlobs map[string][]Blob) error {
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
