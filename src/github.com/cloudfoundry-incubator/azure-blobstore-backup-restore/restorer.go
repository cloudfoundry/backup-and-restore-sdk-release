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

	for sourceContainerId, sourceContainerBackup := range backups {
		destinationContainer := r.containers[sourceContainerId]
		restoreFromStorageAccount, hasRestoreFromStorageAccount := r.restoreFromStorageAccounts[sourceContainerId]

		var err error
		if hasRestoreFromStorageAccount {
			err = destinationContainer.CopyBlobsFromDifferentStorageAccount(restoreFromStorageAccount, sourceContainerBackup.Name, sourceContainerBackup.Blobs)
		} else {
			err = destinationContainer.CopyBlobsFrom(sourceContainerBackup.Name, sourceContainerBackup.Blobs)
		}

		if err != nil {
			return err
		}
	}

	return nil
}
