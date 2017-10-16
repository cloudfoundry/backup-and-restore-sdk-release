package main

import (
	"log"
	"os"

	"github.com/cloudfoundry-incubator/blobstore-backup-restore"
)

func main() {
	awsCliPath := getEnv("AWS_CLI_PATH")
	dropletsBucketName := getEnv("DROPLETS_BUCKET_NAME")
	dropletsBucketRegion := getEnv("DROPLETS_BUCKET_REGION")
	dropletsBucketAccessKey := blobstore.S3AccessKey{
		Id:     getEnv("DROPLETS_BUCKET_ACCESS_KEY_ID"),
		Secret: getEnv("DROPLETS_BUCKET_SECRET_ACCESS_KEY"),
	}

	buildpacksBucketName := getEnv("BUILDPACKS_BUCKET_NAME")
	buildpacksBucketRegion := getEnv("BUILDPACKS_BUCKET_REGION")
	buildpacksBucketAccessKey := blobstore.S3AccessKey{
		Id:     getEnv("BUILDPACKS_BUCKET_ACCESS_KEY_ID"),
		Secret: getEnv("BUILDPACKS_BUCKET_SECRET_ACCESS_KEY"),
	}

	packagesBucketName := getEnv("PACKAGES_BUCKET_NAME")
	packagesBucketRegion := getEnv("PACKAGES_BUCKET_REGION")
	packagesBucketAccessKey := blobstore.S3AccessKey{
		Id:     getEnv("PACKAGES_BUCKET_ACCESS_KEY_ID"),
		Secret: getEnv("PACKAGES_BUCKET_SECRET_ACCESS_KEY"),
	}

	dropletsBucket := blobstore.NewS3Bucket(awsCliPath, dropletsBucketName, dropletsBucketRegion, dropletsBucketAccessKey)
	buildpacksBucket := blobstore.NewS3Bucket(awsCliPath, buildpacksBucketName, buildpacksBucketRegion, buildpacksBucketAccessKey)
	packagesBucket := blobstore.NewS3Bucket(awsCliPath, packagesBucketName, packagesBucketRegion, packagesBucketAccessKey)

	artifact := blobstore.NewFileArtifact(getEnv("BBR_ARTIFACT_DIRECTORY") + "/blobstore.json")

	backuper := blobstore.NewBackuper(dropletsBucket, buildpacksBucket, packagesBucket, artifact)

	err := backuper.Backup()
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
