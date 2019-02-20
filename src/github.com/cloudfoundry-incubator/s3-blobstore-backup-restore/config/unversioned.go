package config

import (
	"fmt"

	"github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/incremental"
	"github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/s3bucket"
)

type UnversionedBucketConfig struct {
	BucketConfig
	Backup BackupBucketConfig `json:"backup"`
}

type BackupBucketConfig struct {
	Name   string `json:"name"`
	Region string `json:"region"`
}

func BuildBackupsToStart(configs map[string]UnversionedBucketConfig) (map[string]incremental.BackupToStart, error) {
	backupsToStart := make(map[string]incremental.BackupToStart)

	for bucketID, config := range configs {
		liveBucket, err := s3bucket.NewBucket(
			config.Name,
			config.Region,
			config.Endpoint,
			s3bucket.AccessKey{
				Id:     config.AwsAccessKeyId,
				Secret: config.AwsSecretAccessKey,
			},
			config.UseIAMProfile,
		)
		if err != nil {
			return nil, err
		}

		backupBucket, err := s3bucket.NewBucket(
			config.Backup.Name,
			config.Backup.Region,
			config.Endpoint,
			s3bucket.AccessKey{
				Id:     config.AwsAccessKeyId,
				Secret: config.AwsSecretAccessKey,
			},
			config.UseIAMProfile,
		)
		if err != nil {
			return nil, err
		}

		backupsToStart[bucketID] = incremental.BackupToStart{
			BucketPair: incremental.BackupBucketPair{
				ConfigLiveBucket:   liveBucket,
				ConfigBackupBucket: backupBucket,
			},
			BackupDirectoryFinder: incremental.Finder{},
		}
	}

	return backupsToStart, nil
}

func BuildBackupsToComplete(
	configs map[string]UnversionedBucketConfig,
	existingBlobsArtifact incremental.Artifact,
) (map[string]incremental.BackupToComplete, error) {
	backupsToComplete := map[string]incremental.BackupToComplete{}

	existingBackups, err := existingBlobsArtifact.Load()
	if err != nil {
		return nil, err
	}

	for bucketID, config := range configs {
		backupBucket, err := s3bucket.NewBucket(
			config.Backup.Name,
			config.Backup.Region,
			config.Endpoint,
			s3bucket.AccessKey{
				Id:     config.AwsAccessKeyId,
				Secret: config.AwsSecretAccessKey,
			},
			config.UseIAMProfile,
		)
		if err != nil {
			return nil, err
		}

		existingBackup, exists := existingBackups[bucketID]
		if !exists {
			return nil, fmt.Errorf("failed to find bucket identifier '%s' in buckets config", bucketID)
		}

		var blobsToCopy []incremental.BackedUpBlob
		for _, path := range existingBackup.Blobs {
			blobsToCopy = append(blobsToCopy, incremental.BackedUpBlob{
				Path:                path,
				BackupDirectoryPath: existingBackup.SrcBackupDirectoryPath,
			})
		}

		backupsToComplete[bucketID] = incremental.BackupToComplete{
			BackupBucket: backupBucket,
			BackupDirectory: incremental.BackupDirectory{
				Path:   existingBackup.DstBackupDirectoryPath,
				Bucket: backupBucket,
			},
			BlobsToCopy: blobsToCopy,
		}
	}

	return backupsToComplete, nil
}

func BuildRestoreBucketPairs(
	configs map[string]UnversionedBucketConfig,
	artifact incremental.Artifact,
) (map[string]incremental.RestoreBucketPair, error) {
	buckets := map[string]incremental.RestoreBucketPair{}

	bucketBackups, err := artifact.Load()
	if err != nil {
		return nil, err
	}

	for bucketID, config := range configs {
		liveBucket, err := s3bucket.NewBucket(
			config.Name,
			config.Region,
			config.Endpoint,
			s3bucket.AccessKey{
				Id:     config.AwsAccessKeyId,
				Secret: config.AwsSecretAccessKey,
			},
			config.UseIAMProfile,
		)
		if err != nil {
			return nil, err
		}

		backupBucket, err := s3bucket.NewBucket(
			bucketBackups[bucketID].BucketName,
			bucketBackups[bucketID].BucketRegion,
			config.Endpoint,
			s3bucket.AccessKey{
				Id:     config.AwsAccessKeyId,
				Secret: config.AwsSecretAccessKey,
			},
			config.UseIAMProfile,
		)
		if err != nil {
			return nil, err
		}

		buckets[bucketID] = incremental.NewRestoreBucketPair(
			liveBucket,
			backupBucket,
		)
	}

	return buckets, nil
}
