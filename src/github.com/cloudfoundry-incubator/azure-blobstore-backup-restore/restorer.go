package azure

import "fmt"

type Restorer struct {
	containers                 map[string]Container
	restoreFromStorageAccounts map[string]StorageAccount
}

func NewRestorer(containers map[string]Container, restoreFromStorageAccounts map[string]StorageAccount) Restorer {
	return Restorer{
		containers:                 containers,
		restoreFromStorageAccounts: restoreFromStorageAccounts,
	}
}

func (r Restorer) Restore(backups map[string]ContainerBackup) error {
	for containerId := range r.containers {
		_, ok := backups[containerId]
		if !ok {
			return fmt.Errorf("container %s is mentioned in the restore config but is not recorded in the artifact", containerId)
		}
	}

	for containerId := range backups {
		destinationContainer, ok := r.containers[containerId]
		if !ok {
			return fmt.Errorf("container %s is not mentioned in the restore config but is present in the artifact", containerId)
		}
		enabled, err := destinationContainer.SoftDeleteEnabled()
		if err != nil {
			return err
		}

		if !enabled {
			return fmt.Errorf("soft delete is not enabled on the given storage account")
		}
	}

	for sourceContainerId, sourceContainerBackup := range backups {
		destinationContainer := r.containers[sourceContainerId]
		restoreFromStorageAccount, hasRestoreFromStorageAccount := r.restoreFromStorageAccounts[sourceContainerId]

		var err error
		if hasRestoreFromStorageAccount {
			err = destinationContainer.CopyBlobsFromDifferentStorageAccount(restoreFromStorageAccount, sourceContainerBackup.Name, sourceContainerBackup.Blobs)
		} else {
			err = destinationContainer.CopyBlobsFromSameStorageAccount(sourceContainerBackup.Name, sourceContainerBackup.Blobs)
		}

		if err != nil {
			return err
		}
	}

	return nil
}
