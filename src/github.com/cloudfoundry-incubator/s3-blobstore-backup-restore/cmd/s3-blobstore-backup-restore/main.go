package main

import (
	"os"

	"encoding/json"
	"io/ioutil"

	"errors"
	"flag"

	"fmt"

	"time"

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

	var runner Runner
	config, err := ioutil.ReadFile(commandFlags.ConfigPath)
	if err != nil {
		exitWithError("Failed to read config: %s", err.Error())
	}

	if commandFlags.Versioned {
		var bucketsConfig map[string]BucketConfig
		err = json.Unmarshal(config, &bucketsConfig)
		if err != nil {
			exitWithError("Failed to parse config: %s", err.Error())
		}

		buckets, err := makeBuckets(bucketsConfig)
		if err != nil {
			exitWithError("Failed to establish session: %s", err.Error())
		}

		artifact := versioned.NewFileArtifact(commandFlags.ArtifactFilePath)

		if commandFlags.IsRestore {
			runner = versioned.NewRestorer(buckets, artifact)
		} else {
			runner = versioned.NewBackuper(buckets, artifact)
		}
	} else {
		var bucketsConfig map[string]BucketConfigWithBackupBucket
		err = json.Unmarshal(config, &bucketsConfig)
		if err != nil {
			exitWithError("Failed to parse config: %s", err.Error())
		}

		bucketPairs, err := makeBucketPairs(bucketsConfig)
		if err != nil {
			exitWithError("Failed to establish session: %s", err.Error())
		}

		artifact := unversioned.NewFileArtifact(commandFlags.ArtifactFilePath)

		if commandFlags.IsRestore {
			runner = unversioned.NewRestorer(bucketPairs, artifact)
		} else {
			runner = unversioned.NewBackuper(bucketPairs, artifact, clock{})
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

func makeBucketPairs(config map[string]BucketConfigWithBackupBucket) (map[string]unversioned.BucketPair, error) {
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

type BackupBucket struct {
	Name   string `json:"name"`
	Region string `json:"region"`
}

type BucketConfigWithBackupBucket struct {
	BucketConfig
	Backup BackupBucket `json:"backup"`
}

func parseFlags() (CommandFlags, error) {
	var configFilePath = flag.String("config", "", "Path to JSON config file")
	var backupAction = flag.Bool("backup", false, "Run blobstore backup")
	var restoreAction = flag.Bool("restore", false, "Run blobstore restore")
	var artifactFilePath = flag.String("artifact-file", "", "Path to the artifact file")
	var unversionedFlag = flag.Bool("unversioned", false, "Indicates targeted buckets are unversioned")

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

	return CommandFlags{
		ConfigPath:       *configFilePath,
		IsRestore:        *restoreAction,
		ArtifactFilePath: *artifactFilePath,
		Versioned:        !*unversionedFlag,
	}, nil
}

type CommandFlags struct {
	ConfigPath       string
	IsRestore        bool
	ArtifactFilePath string
	Versioned        bool
}
