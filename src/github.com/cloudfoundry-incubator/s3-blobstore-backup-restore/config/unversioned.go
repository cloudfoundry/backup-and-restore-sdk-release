package config

import (
	"fmt"

	"github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/unversioned"

	"github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/incremental"
	"github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/s3"
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
		liveBucket, err := s3.NewBucket(
			config.Name,
			config.Region,
			config.Endpoint,
			s3.AccessKey{
				Id:     config.AwsAccessKeyId,
				Secret: config.AwsSecretAccessKey,
			},
			config.UseIAMProfile,
		)
		if err != nil {
			return nil, err
		}

		backupBucket, err := s3.NewBucket(
			config.Backup.Name,
			config.Backup.Region,
			config.Endpoint,
			s3.AccessKey{
				Id:     config.AwsAccessKeyId,
				Secret: config.AwsSecretAccessKey,
			},
			config.UseIAMProfile,
		)
		if err != nil {
			return nil, err
		}

		backupsToStart[bucketID] = incremental.BackupToStart{
			BucketPair: incremental.BucketPair{
				LiveBucket:   liveBucket,
				BackupBucket: backupBucket,
			},
			BackupDirectoryFinder: incremental.Finder{
				Bucket: backupBucket,
			},
		}
	}

	return backupsToStart, nil
}

func BuildBackupsToComplete(
	configs map[string]UnversionedBucketConfig,
	artifact incremental.Artifact,
	existingBlobsArtifact incremental.Artifact,
) (map[string]incremental.BackupToComplete, error) {
	backupsToComplete := map[string]incremental.BackupToComplete{}

	bucketBackups, err := artifact.Load()
	if err != nil {
		return nil, err
	}

	existingBlobsBucketBackups, err := existingBlobsArtifact.Load()
	if err != nil {
		return nil, err
	}

	for bucketID, config := range configs {
		backupBucket, err := s3.NewBucket(
			config.Backup.Name,
			config.Backup.Region,
			config.Endpoint,
			s3.AccessKey{
				Id:     config.AwsAccessKeyId,
				Secret: config.AwsSecretAccessKey,
			},
			config.UseIAMProfile,
		)
		if err != nil {
			return nil, err
		}

		existingBucketBackup, exists := existingBlobsBucketBackups[bucketID]
		if !exists {
			return nil, fmt.Errorf("failed to find bucket identifier '%s' in buckets config", bucketID)
		}

		var blobsToCopy []incremental.BackedUpBlob
		for _, blob := range existingBucketBackup.Blobs {
			blobsToCopy = append(blobsToCopy, incremental.BackedUpBlob{
				Path:                blob,
				BackupDirectoryPath: existingBucketBackup.BackupDirectoryPath,
			})
		}

		backupsToComplete[bucketID] = incremental.BackupToComplete{
			BackupBucket: backupBucket,
			BackupDirectory: incremental.BackupDirectory{
				Path:   bucketBackups[bucketID].BackupDirectoryPath,
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
) (map[string]unversioned.RestoreBucketPair, error) {
	buckets := map[string]unversioned.RestoreBucketPair{}

	bucketBackups, err := artifact.Load()
	if err != nil {
		return nil, err
	}

	for bucketID, config := range configs {
		liveBucket, err := s3.NewBucket(
			config.Name,
			config.Region,
			config.Endpoint,
			s3.AccessKey{
				Id:     config.AwsAccessKeyId,
				Secret: config.AwsSecretAccessKey,
			},
			config.UseIAMProfile,
		)
		if err != nil {
			return nil, err
		}

		backupBucket, err := s3.NewBucket(
			bucketBackups[bucketID].BucketName,
			bucketBackups[bucketID].BucketRegion,
			config.Endpoint,
			s3.AccessKey{
				Id:     config.AwsAccessKeyId,
				Secret: config.AwsSecretAccessKey,
			},
			config.UseIAMProfile,
		)
		if err != nil {
			return nil, err
		}

		buckets[bucketID] = unversioned.NewRestoreBucketPair(
			liveBucket,
			backupBucket,
		)
	}

	return buckets, nil
}
