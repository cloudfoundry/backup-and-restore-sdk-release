package main

import (
	"os"

	"encoding/json"
	"io/ioutil"

	"errors"
	"flag"

	"fmt"

	"github.com/cloudfoundry-incubator/blobstore-backup-restore"
)

func main() {
	commandFlags, err := parseFlags()
	if err != nil {
		exitWithError(err.Error())
	}

	versionedArtifact := blobstore.NewVersionedFileArtifact(commandFlags.ArtifactFilePath)

	config, err := ioutil.ReadFile(commandFlags.ConfigPath)
	if err != nil {
		exitWithError("Failed to read config: %s", err.Error())
	}

	var bucketsConfig map[string]BucketConfig
	err = json.Unmarshal(config, &bucketsConfig)
	if err != nil {
		exitWithError("Failed to parse config: %s", err.Error())
	}

	buckets, err := makeBuckets(bucketsConfig)
	if err != nil {
		exitWithError("Failed to establish session: %s", err.Error())
	}

	var runner blobstore.Runner
	if commandFlags.IsRestore {
		runner = blobstore.NewVersionedRestorer(buckets, versionedArtifact)
	} else {
		runner = blobstore.NewVersionedBackuper(buckets, versionedArtifact)
	}

	err = runner.Run()
	if err != nil {
		exitWithError(err.Error())
	}
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

func makeBuckets(config map[string]BucketConfig) (map[string]blobstore.Bucket, error) {
	var buckets = map[string]blobstore.Bucket{}

	var err error
	for identifier, bucketConfig := range config {
		buckets[identifier], err = blobstore.NewS3Bucket(
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

type BucketConfig struct {
	Name               string `json:"name"`
	Region             string `json:"region"`
	AwsAccessKeyId     string `json:"aws_access_key_id"`
	AwsSecretAccessKey string `json:"aws_secret_access_key"`
	Endpoint           string `json:"endpoint"`
}

func parseFlags() (CommandFlags, error) {
	var configFilePath = flag.String("config", "", "Path to JSON config file")
	var backupAction = flag.Bool("backup", false, "Run blobstore backup")
	var restoreAction = flag.Bool("restore", false, "Run blobstore restore")
	var artifactFilePath = flag.String("artifact-file", "", "Path to the artifact file")

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
	}, nil
}

type CommandFlags struct {
	ConfigPath       string
	IsRestore        bool
	ArtifactFilePath string
}
