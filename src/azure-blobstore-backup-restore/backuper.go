package azure

import "fmt"

type BlobId struct {
	Name string `json:"name"`
	ETag string `json:"etag"`
}

type ContainerBackup struct {
	Name  string   `json:"name"`
	Blobs []BlobId `json:"blobs"`
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
		enabled, err := container.SoftDeleteEnabled()
		if err != nil {
			return nil, err
		}

		if !enabled {
			return nil, fmt.Errorf("soft delete is not enabled on the given storage account")
		}

		blobs, err := container.ListBlobs()
		if err != nil {
			return nil, err
		}

		backups[containerId] = ContainerBackup{Name: container.Name(), Blobs: blobs}
	}

	return backups, nil
}
