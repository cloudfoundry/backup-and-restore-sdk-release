package main

import (
	"flag"
	"log"

	"github.com/cloudfoundry-incubator/azure-blobstore-backup-restore"
)

func main() {
	var artifactFilePath = flag.String("artifact-file", "", "Path to the artifact file")
	var configFilePath = flag.String("config", "", "Path to JSON config file")
	var backupAction = flag.Bool("backup", false, "Run blobstore backup")
	var restoreAction = flag.Bool("restore", false, "Run blobstore restore")

	flag.Parse()

	if *backupAction && *restoreAction {
		log.Fatalf("only one of: --backup or --restore can be provided")
	}

	if !*backupAction && !*restoreAction {
		log.Fatalf("missing --backup or --restore flag")
	}

	config, err := azure.ParseConfig(*configFilePath)
	exitOnError(err)

	containerBuilder := azure.NewContainerBuilder(config)
	containers, err := containerBuilder.Containers()
	exitOnError(err)

	artifact := azure.NewArtifact(*artifactFilePath)

	if *backupAction {
		backuper := azure.NewBackuper(containers)

		backups, err := backuper.Backup()
		exitOnError(err)

		err = artifact.Write(backups)
		exitOnError(err)
	} else {
		restoreFromStorageAccounts := containerBuilder.RestoreFromStorageAccounts()

		restorer := azure.NewRestorer(containers, restoreFromStorageAccounts)

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
