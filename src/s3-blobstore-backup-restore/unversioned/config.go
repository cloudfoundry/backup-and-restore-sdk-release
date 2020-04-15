package unversioned

import (
	"fmt"

	"s3-blobstore-backup-restore/incremental"
	"s3-blobstore-backup-restore/s3bucket"
)

type UnversionedBucketConfig struct {
	Name               string             `json:"name"`
	Region             string             `json:"region"`
	AwsAccessKeyId     string             `json:"aws_access_key_id"`
	AwsSecretAccessKey string             `json:"aws_secret_access_key"`
	Endpoint           string             `json:"endpoint"`
	UseIAMProfile      bool               `json:"use_iam_profile"`
	Backup             BackupBucketConfig `json:"backup"`
	ForcePathStyle			 bool								`json:"force_path_style"`
}

type BackupBucketConfig struct {
	Name   string `json:"name"`
	Region string `json:"region"`
}

type NewBucket func(bucketName, bucketRegion, endpoint string, accessKey s3bucket.AccessKey, useIAMProfile, usePathStyle bool) (Bucket, error)

func BuildBackupsToStart(configs map[string]UnversionedBucketConfig, newBucket NewBucket) (map[string]incremental.BackupToStart, error) {
	backupsToStart := make(map[string]incremental.BackupToStart)

	for bucketID, config := range configs {
		liveBucket, err := newBucket(
			config.Name,
			config.Region,
			config.Endpoint,
			s3bucket.AccessKey{
				Id:     config.AwsAccessKeyId,
				Secret: config.AwsSecretAccessKey,
			},
			config.UseIAMProfile,
			config.ForcePathStyle,
		)
		if err != nil {
			return nil, err
		}

		if err := bucketIsVersioned(liveBucket); err != nil {
			return nil, err
		}

		backupBucket, err := newBucket(
			config.Backup.Name,
			config.Backup.Region,
			config.Endpoint,
			s3bucket.AccessKey{
				Id:     config.AwsAccessKeyId,
				Secret: config.AwsSecretAccessKey,
			},
			config.UseIAMProfile,
			config.ForcePathStyle,
		)
		if err != nil {
			return nil, err
		}

		if err := bucketIsVersioned(backupBucket); err != nil {
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

	backupsToStart = incremental.MarkSameBackupsToStart(backupsToStart)

	return backupsToStart, nil
}

func bucketIsVersioned(bucket Bucket) error {
	isVersioned, err := bucket.IsVersioned()
	if err != nil {
		return err
	}

	if isVersioned {
		return fmt.Errorf("bucket %s is versioned", bucket.Name())
	}
	return nil
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
		existingBackup, exists := existingBackups[bucketID]
		if !exists {
			return nil, fmt.Errorf("failed to find bucket identifier '%s' in buckets config", bucketID)
		}

		if existingBackup.SameBucketAs != "" {
			backupsToComplete[bucketID] = incremental.BackupToComplete{
				SameAsBucketID: existingBackup.SameBucketAs,
			}
			continue
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
			s3bucket.ForcePathStyleDuringTheRefactor,
		)
		if err != nil {
			return nil, err
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
	newBucket NewBucket,
) (map[string]incremental.RestoreBucketPair, error) {
	pairs := map[string]incremental.RestoreBucketPair{}

	backups, err := artifact.Load()
	if err != nil {
		return nil, err
	}

	for bucketID, config := range configs {
		if _, ok := backups[bucketID]; !ok {
			return nil, fmt.Errorf("backup artifact does not contain bucket ID '%s'", bucketID)
		}

		if backups[bucketID].SameBucketAs != "" {
			pairs[bucketID] = incremental.RestoreBucketPair{
				SameAsBucketID: backups[bucketID].SameBucketAs,
			}
			continue
		}

		liveBucket, err := newBucket(
			config.Name,
			config.Region,
			config.Endpoint,
			s3bucket.AccessKey{
				Id:     config.AwsAccessKeyId,
				Secret: config.AwsSecretAccessKey,
			},
			config.UseIAMProfile,
			s3bucket.ForcePathStyleDuringTheRefactor,
		)

		if err != nil {
			return nil, err
		}

		if err := bucketIsVersioned(liveBucket); err != nil {
			return nil, err
		}

		backupBucket, err := newBucket(
			backups[bucketID].BucketName,
			backups[bucketID].BucketRegion,
			config.Endpoint,
			s3bucket.AccessKey{
				Id:     config.AwsAccessKeyId,
				Secret: config.AwsSecretAccessKey,
			},
			config.UseIAMProfile,
			s3bucket.ForcePathStyleDuringTheRefactor,
		)

		if err != nil {
			return nil, err
		}

		if err := bucketIsVersioned(backupBucket); err != nil {
			return nil, err
		}

		pairs[bucketID] = incremental.RestoreBucketPair{
			ConfigLiveBucket:     liveBucket,
			ArtifactBackupBucket: backupBucket,
		}
	}

	return pairs, nil
}
