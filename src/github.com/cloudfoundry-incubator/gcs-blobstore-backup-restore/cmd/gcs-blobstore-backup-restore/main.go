package main

import (
	"flag"
	"log"

	"github.com/cloudfoundry-incubator/gcs-blobstore-backup-restore"
)

func main() {
	artifactPath := flag.String("artifact-file", "", "Path to the artifact file")
	configPath := flag.String("config", "", "Path to JSON config file")
	_ = flag.Bool("backup", false, "Run blobstore backup")

	flag.Parse()

	config, err := gcs.ParseConfig(*configPath)
	exitOnError(err)

	buckets, err := gcs.BuildBuckets(config)
	exitOnError(err)

	backuper := gcs.NewBackuper(buckets)

	backups, err := backuper.Backup()
	exitOnError(err)

	artifact := gcs.NewArtifact(*artifactPath)
	err = artifact.Write(backups)
	exitOnError(err)
}

func exitOnError(err error) {
	if err != nil {
		log.Fatalf(err.Error())
	}
}
