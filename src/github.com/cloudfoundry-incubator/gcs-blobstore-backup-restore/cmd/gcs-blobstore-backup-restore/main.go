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

	flag.Parse()

	if !*backupAction && !*restoreAction {
		log.Fatal("missing --backup or --restore flag")
	}

	if *backupAction && *restoreAction {
		log.Fatal("only one of: --backup or --restore can be provided")
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
