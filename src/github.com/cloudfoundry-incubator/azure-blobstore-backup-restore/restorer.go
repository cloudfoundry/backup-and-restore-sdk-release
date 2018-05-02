package azure

import "fmt"

type Restorer struct {
	containers map[string]Container
}

func NewRestorer(containers map[string]Container) Restorer {
	return Restorer{containers: containers}
}

func (r Restorer) Restore(backups map[string]ContainerBackup) error {
	for containerId := range backups {
		destinationContainer := r.containers[containerId]
		enabled, err := destinationContainer.SoftDeleteEnabled()
		if err != nil {
			return err
		}

		if !enabled {
			return fmt.Errorf("soft delete is not enabled on the given storage account")
		}
	}

	for containerId, sourceContainerBackup := range backups {
		destinationContainer := r.containers[containerId]
		for _, blob := range sourceContainerBackup.Blobs {
			err := destinationContainer.CopyFrom(sourceContainerBackup.Name, blob.Name, blob.Etag)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
