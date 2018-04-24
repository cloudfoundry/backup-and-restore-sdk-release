package azure

import "fmt"

type Blob struct {
	Name string `json:"name"`
	Hash string `json:"hash"`
}

type ContainerBackup struct {
	Name  string `json:"name"`
	Blobs []Blob `json:"blobs"`
}

type Backuper struct {
	containers map[string]Container
}

func NewBackuper(containers map[string]Container) Backuper {
	return Backuper{containers: containers}
}

func (b Backuper) Backup() (map[string]ContainerBackup, error) {
	var backups = make(map[string]ContainerBackup)

	for containerId, container := range b.containers {
		disabled, err := container.SoftDeleteIsDisabled()
		if err != nil {
			return nil, err
		}

		if disabled {
			return nil, fmt.Errorf("soft delete is not enabled on container: '%s'", container.Name())
		}

		blobs, err := container.ListBlobs()
		if err != nil {
			return nil, err
		}

		backups[containerId] = ContainerBackup{Name: container.Name(), Blobs: blobs}
	}

	return backups, nil
}
