package main

import (
	"log"
	"os"

	"encoding/json"
	"io/ioutil"

	"github.com/cloudfoundry-incubator/blobstore-backup-restore"
	"flag"
	"errors"
)

func main() {
	configFilePath, artifactFilePath, err := parseFlags()
	if err != nil {
		log.Fatal(err.Error())
	}

	awsCliPath := getEnv("AWS_CLI_PATH")

	artifact := blobstore.NewFileArtifact(artifactFilePath)

	config, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		log.Fatal("Failed to read config")
	}

	var bucketsConfig map[string]BucketConfig
	err = json.Unmarshal(config, &bucketsConfig)
	if err != nil {
		log.Fatal("Failed to parse config")
	}

	buckets := makeBuckets(awsCliPath, bucketsConfig)

	backuper := blobstore.NewBackuper(buckets, artifact)

	err = backuper.Backup()
	if err != nil {
		log.Fatal(err.Error())
	}
}

func getEnv(varName string) string {
	value, exists := os.LookupEnv(varName)
	if !exists {
		log.Fatalf("Missing environment variable '%s'", varName)
	}
	return value
}

func makeBuckets(awsCliPath string, config map[string]BucketConfig) []blobstore.Bucket {
	var buckets = []blobstore.Bucket{}

	for identifier, bucketConfig := range config {
		buckets = append(buckets, blobstore.NewS3Bucket(
			awsCliPath,
			identifier,
			bucketConfig.Name,
			bucketConfig.Region,
			blobstore.S3AccessKey{
				Id:     bucketConfig.AwsAccessKeyId,
				Secret: bucketConfig.AwsSecretAccessKey,
			},
		))
	}

	return buckets
}

type BucketConfig struct {
	Name               string `json:"name"`
	Region             string `json:"region"`
	AwsAccessKeyId     string `json:"aws_access_key_id"`
	AwsSecretAccessKey string `json:"aws_secret_access_key"`
}

func parseFlags() (string, string, error) {
	var configFilePath = flag.String("config", "", "Path to JSON config file")
	var artifactFilePath = flag.String("artifact-file", "", "Path to output file")

	flag.Parse()

	if *configFilePath == "" {
		return "", "", errors.New("missing --config flag")
	}

	if *artifactFilePath == "" {
		return "", "", errors.New("missing --artifact-file flag")
	}

	return *configFilePath, *artifactFilePath, nil
}