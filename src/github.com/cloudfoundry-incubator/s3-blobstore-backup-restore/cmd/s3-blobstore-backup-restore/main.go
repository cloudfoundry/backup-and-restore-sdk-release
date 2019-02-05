package main

import (
	"os"

	"encoding/json"
	"io/ioutil"

	"errors"
	"flag"

	"fmt"

	"time"

	"github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/incremental"
	"github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/s3"
	"github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/unversioned"
	"github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/versioned"
)

type Runner interface {
	Run() error
}

func main() {
	commandFlags, err := parseFlags()
	if err != nil {
		exitWithError(err.Error())
	}

	config, err := ioutil.ReadFile(commandFlags.ConfigPath)
	if err != nil {
		exitWithError(fmt.Sprintf("Failed to read config: %s", err.Error()))
	}

	var runner Runner
	if commandFlags.Versioned {
		var bucketsConfig map[string]BucketConfig
		err = json.Unmarshal(config, &bucketsConfig)
		if err != nil {
			exitWithError(fmt.Sprintf("Failed to parse config: %s", err.Error()))
		}

		buckets, err := makeBuckets(bucketsConfig)
		if err != nil {
			exitWithError(fmt.Sprintf("Failed to establish session: %s", err.Error()))
		}

		artifact := versioned.NewFileArtifact(commandFlags.ArtifactFilePath)

		if commandFlags.IsRestore {
			runner = versioned.NewRestorer(buckets, artifact)
		} else {
			runner = versioned.NewBackuper(buckets, artifact)
		}
	} else {
		var bucketsConfig map[string]UnversionedBucketConfig
		err = json.Unmarshal(config, &bucketsConfig)
		if err != nil {
			exitWithError(fmt.Sprintf("Failed to parse config: %s", err.Error()))
		}

		switch {
		case commandFlags.IsRestore:
			{
				artifact := unversioned.NewFileArtifact(commandFlags.ArtifactFilePath)

				bucketPairs, err := makeBucketPairs(bucketsConfig)
				if err != nil {
					exitWithError(fmt.Sprintf("Failed to establish session: %s", err.Error()))
				}

				runner = unversioned.NewRestorer(bucketPairs, artifact)
			}
		case commandFlags.UnversionedCompleter:
			{
				existingBackupBlobsArtifact := incremental.NewArtifact(commandFlags.ExistingBackupBlobsArtifactFilePath)
				backupArtifact := incremental.NewArtifact(commandFlags.ArtifactFilePath)
				backupsToComplete, err := makeIncrementalBackupsToComplete(bucketsConfig, backupArtifact, existingBackupBlobsArtifact)
				if err != nil {
					exitWithError(fmt.Sprintf("Failed to deserialise incremental backups to complete: %s", err.Error()))
				}
				runner = incremental.BackupCompleter{
					BackupsToComplete: backupsToComplete,
				}
			}
		default:
			{
				backupsToStart, err := makeIncrementalBackupsToStart(bucketsConfig)
				if err != nil {
					exitWithError(fmt.Sprintf("Failed to deserialise incremental backups to start: %s", err.Error()))
				}
				backupArtifact := incremental.NewArtifact(commandFlags.ArtifactFilePath)
				existingBackupBlobsArtifact := incremental.NewArtifact(commandFlags.ExistingBackupBlobsArtifactFilePath)
				runner = incremental.NewBackupStarter(backupsToStart, clock{}, backupArtifact, existingBackupBlobsArtifact)
			}
		}
	}

	err = runner.Run()
	if err != nil {
		exitWithError(err.Error())
	}
}

type clock struct {
}

func (c clock) Now() string {
	return time.Now().Format("2006_01_02_15_04_05")
}

func exitWithError(a ...interface{}) {
	fmt.Fprintln(os.Stderr, a...)
	os.Exit(1)
}

func makeBuckets(config map[string]BucketConfig) (map[string]s3.VersionedBucket, error) {
	var buckets = map[string]s3.VersionedBucket{}

	for identifier, bucketConfig := range config {
		s3Bucket, err := s3.NewBucket(
			bucketConfig.Name,
			bucketConfig.Region,
			bucketConfig.Endpoint,
			s3.AccessKey{
				Id:     bucketConfig.AwsAccessKeyId,
				Secret: bucketConfig.AwsSecretAccessKey,
			},
			bucketConfig.UseIAMProfile,
		)
		if err != nil {
			return nil, err
		}

		buckets[identifier] = s3Bucket
	}

	return buckets, nil
}

func makeIncrementalBackupsToStart(config map[string]UnversionedBucketConfig) (map[string]incremental.BackupsToStart, error) {
	var buckets = map[string]incremental.BackupsToStart{}

	for identifier, bucketConfig := range config {
		liveBucket, err := s3.NewBucket(
			bucketConfig.Name,
			bucketConfig.Region,
			bucketConfig.Endpoint,
			s3.AccessKey{
				Id:     bucketConfig.AwsAccessKeyId,
				Secret: bucketConfig.AwsSecretAccessKey,
			},
			bucketConfig.UseIAMProfile,
		)
		if err != nil {
			return nil, err
		}

		backupBucket, err := s3.NewBucket(
			bucketConfig.Backup.Name,
			bucketConfig.Backup.Region,
			bucketConfig.Endpoint,
			s3.AccessKey{
				Id:     bucketConfig.AwsAccessKeyId,
				Secret: bucketConfig.AwsSecretAccessKey,
			},
			bucketConfig.UseIAMProfile,
		)
		if err != nil {
			return nil, err
		}

		bucketPair := incremental.BucketPair{
			LiveBucket:   liveBucket,
			BackupBucket: backupBucket,
		}

		buckets[identifier] = incremental.BackupsToStart{
			BucketPair: bucketPair,
			BackupDirectoryFinder: incremental.Finder{
				ID:     identifier,
				Bucket: backupBucket,
			},
		}
	}

	return buckets, nil
}

func makeIncrementalBackupsToComplete(config map[string]UnversionedBucketConfig, backupArtifact, existingBlobsArtifact incremental.Artifact) (map[string]incremental.BackupToComplete, error) {
	var backupsToComplete = map[string]incremental.BackupToComplete{}

	existingBucketBackups, _ := existingBlobsArtifact.Load()
	bucketBackups, _ := backupArtifact.Load()

	for identifier, bucketConfig := range config {
		existingBucketBackup := existingBucketBackups[identifier]
		bucketBackup := bucketBackups[identifier]

		backupBucket, err := s3.NewBucket(
			bucketBackup.BucketName,
			bucketConfig.Backup.Region,
			bucketConfig.Endpoint,
			s3.AccessKey{
				Id:     bucketConfig.AwsAccessKeyId,
				Secret: bucketConfig.AwsSecretAccessKey,
			},
			bucketConfig.UseIAMProfile,
		)
		if err != nil {
			return nil, err
		}

		var blobsToCopy []incremental.BackedUpBlob

		for _, blobPath := range existingBucketBackup.Blobs {
			blobsToCopy = append(blobsToCopy, incremental.BackedUpBlob{Path: blobPath, BackupDirectoryPath: existingBucketBackup.BackupDirectoryPath})
		}

		backupsToComplete[identifier] = incremental.BackupToComplete{
			BackupBucket: backupBucket,
			BackupDirectory: incremental.BackupDirectory{
				Path:   bucketBackup.BackupDirectoryPath,
				Bucket: backupBucket,
			},
			BlobsToCopy: blobsToCopy,
		}
	}

	return backupsToComplete, nil
}

func makeBucketPairs(config map[string]UnversionedBucketConfig) (map[string]unversioned.BucketPair, error) {
	var buckets = map[string]unversioned.BucketPair{}

	for identifier, bucketConfig := range config {
		liveBucket, err := s3.NewBucket(
			bucketConfig.Name,
			bucketConfig.Region,
			bucketConfig.Endpoint,
			s3.AccessKey{
				Id:     bucketConfig.AwsAccessKeyId,
				Secret: bucketConfig.AwsSecretAccessKey,
			},
			bucketConfig.UseIAMProfile,
		)
		if err != nil {
			return nil, err
		}

		backupBucket, err := s3.NewBucket(
			bucketConfig.Backup.Name,
			bucketConfig.Backup.Region,
			bucketConfig.Endpoint,
			s3.AccessKey{
				Id:     bucketConfig.AwsAccessKeyId,
				Secret: bucketConfig.AwsSecretAccessKey,
			},
			bucketConfig.UseIAMProfile,
		)
		if err != nil {
			return nil, err
		}

		buckets[identifier] = unversioned.NewS3BucketPair(
			liveBucket,
			backupBucket,
		)
	}

	return buckets, nil
}

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

func parseFlags() (CommandFlags, error) {
	var configFilePath = flag.String("config", "", "Path to JSON config file")
	var backupAction = flag.Bool("backup", false, "Run blobstore backup")
	var restoreAction = flag.Bool("restore", false, "Run blobstore restore")
	var artifactFilePath = flag.String("artifact-file", "", "Path to the artifact file")
	var existingBackupBlobsArtifactFilePath = flag.String("existing-backup-blobs-artifact", "", "Path to the existing backup blobs artifact file")
	var unversionedBackupStarter = flag.Bool("unversioned-backup-starter", false, "Run backup starter for unversioned buckets")
	var unversionedBackupCompleter = flag.Bool("unversioned-backup-completer", false, "Run backup completer for unversioned buckets")

	flag.Parse()

	if *backupAction && *restoreAction {
		return CommandFlags{}, errors.New("only one of: --backup or --restore can be provided")
	}

	if !*backupAction && !*restoreAction {
		return CommandFlags{}, errors.New("missing --backup or --restore flag")
	}

	if *configFilePath == "" {
		return CommandFlags{}, errors.New("missing --config flag")
	}

	if *artifactFilePath == "" {
		return CommandFlags{}, errors.New("missing --artifact-file flag")
	}

	if *unversionedBackupStarter && *unversionedBackupCompleter {
		return CommandFlags{}, errors.New("at most one of: --unversioned-backup-starter or --unversioned-backup-completer can be provided")
	}

	if *backupAction && (*unversionedBackupStarter || *unversionedBackupCompleter) && *existingBackupBlobsArtifactFilePath == "" {
		return CommandFlags{}, errors.New("missing --existing-backup-blobs-artifact")
	}

	return CommandFlags{
		ConfigPath:                          *configFilePath,
		IsRestore:                           *restoreAction,
		ArtifactFilePath:                    *artifactFilePath,
		ExistingBackupBlobsArtifactFilePath: *existingBackupBlobsArtifactFilePath,
		Versioned:                           !*unversionedBackupStarter && !*unversionedBackupCompleter,
		UnversionedCompleter:                *unversionedBackupCompleter,
	}, nil
}

type CommandFlags struct {
	ConfigPath                          string
	IsRestore                           bool
	ArtifactFilePath                    string
	ExistingBackupBlobsArtifactFilePath string
	Versioned                           bool
	UnversionedCompleter                bool
}
