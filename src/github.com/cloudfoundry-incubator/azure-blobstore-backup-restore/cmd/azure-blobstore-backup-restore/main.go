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

	containers, err := buildContainers(config)
	exitOnError(err)

	artifact := azure.NewArtifact(*artifactFilePath)

	if *backupAction {
		backuper := azure.NewBackuper(containers)

		backups, err := backuper.Backup()
		exitOnError(err)

		err = artifact.Write(backups)
		exitOnError(err)
	} else {
		restorer := azure.NewRestorer(containers)

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

func buildContainers(config map[string]azure.ContainerConfig) (map[string]azure.Container, error) {
	var containers = make(map[string]azure.Container)

	for containerId, containerConfig := range config {
		container, err := azure.NewContainer(
			containerConfig.Name,
			containerConfig.StorageAccount,
			containerConfig.StorageKey,
		)
		if err != nil {
			return nil, err
		}

		containers[containerId] = container
	}

	return containers, nil
}
