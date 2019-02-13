package config

import (
	"github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/incremental"
	"github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/s3"
)

type BucketConfig struct {
	Name               string `json:"name"`
	Region             string `json:"region"`
	AwsAccessKeyId     string `json:"aws_access_key_id"`
	AwsSecretAccessKey string `json:"aws_secret_access_key"`
	Endpoint           string `json:"endpoint"`
	UseIAMProfile      bool   `json:"use_iam_profile"`
}

type BackupBucketConfig struct {
	Name   string `json:"name"`
	Region string `json:"region"`
}

type UnversionedBucketConfig struct {
	BucketConfig
	Backup BackupBucketConfig `json:"backup"`
}

func BuildIncrementalBackupsToStart(configs map[string]UnversionedBucketConfig) (map[string]incremental.BackupToStart, error) {
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
