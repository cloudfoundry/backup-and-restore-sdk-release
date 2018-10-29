package main

import (
	"flag"
	"log"

	"github.com/cloudfoundry-incubator/gcs-blobstore-backup-restore"
)

func main() {
	artifactPath := flag.String("artifact-file", "", "Path to the artifact file")
	configPath := flag.String("config", "", "Path to JSON config file")
	gcpServiceAccountKeyPath := flag.String("gcp-service-account-key", "", "Path to GCP service account key")
	backupAction := flag.Bool("backup", false, "Run blobstore backup")
	restoreAction := flag.Bool("restore", false, "Run blobstore restore")

	flag.Parse()

	if !*backupAction && !*restoreAction {
		log.Fatal("missing --backup or --restore flag")
	}

	if *backupAction && *restoreAction {
		log.Fatal("only one of: --backup or --restore can be provided")
	}

	config, err := gcs.ParseConfig(*configPath)
	exitOnError(err)

	gcpServiceAccountKey, err := gcs.ReadGCPServiceAccountKey(*gcpServiceAccountKeyPath)
	exitOnError(err)

	buckets, err := gcs.BuildBuckets(gcpServiceAccountKey, config)
	exitOnError(err)

	artifact := gcs.NewArtifact(*artifactPath)

	if *backupAction {
		backuper := gcs.NewBackuper(buckets)

		backupBucketDirectories, commonBlobs, err := backuper.CreateLiveBucketSnapshot()
		exitOnError(err)

		err = backuper.CopyBlobsWithinBackupBucket(backupBucketDirectories, commonBlobs)
		exitOnError(err)

		err = artifact.Write(backupBucketDirectories)
		exitOnError(err)
	} else {
		restorer := gcs.NewRestorer(buckets)

		backupBuckets, err := artifact.Read()
		exitOnError(err)

		err = restorer.Restore(backupBuckets)
		exitOnError(err)
	}
}

func exitOnError(err error) {
	if err != nil {
		log.Fatalf(err.Error())
	}
}
