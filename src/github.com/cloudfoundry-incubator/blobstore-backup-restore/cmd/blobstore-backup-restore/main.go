package main

import (
	"os"

	"encoding/json"
	"io/ioutil"

	"errors"
	"flag"

	"fmt"

	"time"

	"github.com/cloudfoundry-incubator/blobstore-backup-restore"
)

func main() {
	commandFlags, err := parseFlags()
	if err != nil {
		exitWithError(err.Error())
	}

	var runner blobstore.Runner
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

		artifact := blobstore.NewVersionedFileArtifact(commandFlags.ArtifactFilePath)

		if commandFlags.IsRestore {
			runner = blobstore.NewVersionedRestorer(buckets, artifact)
		} else {
			runner = blobstore.NewVersionedBackuper(buckets, artifact)
		}
	} else {
		var bucketsConfig map[string]BucketConfigWithBackupBucket
		err = json.Unmarshal(config, &bucketsConfig)
		if err != nil {
			exitWithError("Failed to parse config: %s", err.Error())
		}

		buckets, err := makeBucketPairs(bucketsConfig)
		if err != nil {
			exitWithError("Failed to establish session: %s", err.Error())
		}

		artifact := blobstore.NewUnversionedFileArtifact(commandFlags.ArtifactFilePath)

		if commandFlags.IsRestore {
			runner = blobstore.NewUnversionedRestorer(buckets, artifact)
		} else {
			runner = blobstore.NewUnversionedBackuper(buckets, artifact, clock{})
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

func getEnv(varName string) string {
	value, exists := os.LookupEnv(varName)
	if !exists {
		exitWithError("Missing environment variable '%s'", varName)
	}
	return value
}

func makeBuckets(config map[string]BucketConfig) (map[string]blobstore.VersionedBucket, error) {
	var buckets = map[string]blobstore.VersionedBucket{}

	var err error
	for identifier, bucketConfig := range config {
		buckets[identifier], err = blobstore.NewS3VersionedBucket(
			bucketConfig.Name,
			bucketConfig.Region,
			bucketConfig.Endpoint,
			blobstore.S3AccessKey{
				Id:     bucketConfig.AwsAccessKeyId,
				Secret: bucketConfig.AwsSecretAccessKey,
			},
		)

		if err != nil {
			return nil, err
		}
	}

	return buckets, nil
}

func makeBucketPairs(config map[string]BucketConfigWithBackupBucket) (map[string]blobstore.UnversionedBucketPair, error) {
	var buckets = map[string]blobstore.UnversionedBucketPair{}

	var err error
	for identifier, bucketConfig := range config {
		buckets[identifier], err = blobstore.NewS3BucketPair(
			bucketConfig.Name,
			bucketConfig.Region,
			bucketConfig.Endpoint,
			blobstore.S3AccessKey{
				Id:     bucketConfig.AwsAccessKeyId,
				Secret: bucketConfig.AwsSecretAccessKey,
			},
			bucketConfig.Backup.Name,
			bucketConfig.Backup.Region,
		)

		if err != nil {
			return nil, err
		}
	}

	return buckets, nil
}

type BucketConfig struct {
	Name               string `json:"name"`
	Region             string `json:"region"`
	AwsAccessKeyId     string `json:"aws_access_key_id"`
	AwsSecretAccessKey string `json:"aws_secret_access_key"`
	Endpoint           string `json:"endpoint"`
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
