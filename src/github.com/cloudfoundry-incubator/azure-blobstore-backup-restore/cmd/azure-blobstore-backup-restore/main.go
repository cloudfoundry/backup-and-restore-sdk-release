package main

import (
	"flag"
	"log"

	"github.com/cloudfoundry-incubator/azure-blobstore-backup-restore"
)

func main() {
	var artifactFilePath = flag.String("artifact-file", "", "Path to the artifact file")
	var configFilePath = flag.String("config", "", "Path to JSON config file")
	var _ = flag.Bool("backup", false, "Run blobstore backup")

	flag.Parse()

	config, err := azure.ParseConfig(*configFilePath)
	exitOnError(err)

	containers, err := buildContainers(config)
	exitOnError(err)

	backuper := azure.NewBackuper(containers)
	backups, err := backuper.Backup()
	exitOnError(err)

	artifact := azure.NewArtifact(*artifactFilePath)
	err = artifact.Write(backups)
	exitOnError(err)
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
			containerConfig.AzureAccountName,
			containerConfig.AzureAccountKey,
		)
		if err != nil {
			return nil, err
		}

		containers[containerId] = container
	}

	return containers, nil
}
