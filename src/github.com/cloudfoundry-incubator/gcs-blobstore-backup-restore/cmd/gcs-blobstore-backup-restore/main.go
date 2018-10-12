package main

import (
	"flag"
	"log"

	"github.com/cloudfoundry-incubator/gcs-blobstore-backup-restore"
)

func main() {
	artifactPath := flag.String("artifact-file", "", "Path to the artifact file")
	configPath := flag.String("config", "", "Path to JSON config file")
	backupAction := flag.Bool("backup", false, "Run blobstore backup")
	restoreAction := flag.Bool("restore", false, "Run blobstore restore")
	unlockAction := flag.Bool("unlock", false, "Run blobstore unlock")

	flag.Parse()

	if !*backupAction && !*restoreAction && !*unlockAction {
		log.Fatal("missing --unlock, --backup or --restore flag")
	}

	if (*backupAction && *restoreAction) || (*backupAction && *unlockAction) || (*unlockAction && *restoreAction) {
		log.Fatal("only one of: --unlock, --backup or --restore can be provided")
	}

	config, err := gcs.ParseConfig(*configPath)
	exitOnError(err)

	buckets, err := gcs.BuildBuckets(config)
	exitOnError(err)

	artifact := gcs.NewArtifact(*artifactPath)

	executionStrategy := gcs.NewParallelStrategy()

	if *backupAction {
		backuper := gcs.NewBackuper(buckets)

		err := backuper.CreateLiveBucketSnapshot()
		exitOnError(err)

		//err = artifact.Write(buckets)
		//exitOnError(err)
	} else if *unlockAction {
		backuper := gcs.NewBackuper(buckets)

		err := backuper.TransferBlobsToBackupBucket()
		exitOnError(err)
	} else {
		restorer := gcs.NewRestorer(buckets, executionStrategy)

		backups, err := artifact.Read()
		exitOnError(err)

		err = restorer.Restore(backups)
		exitOnError(err)
	}
}

func exitOnError(err error) {
	if err != nil {
		log.Fatalf(err.Error())
	}
}
