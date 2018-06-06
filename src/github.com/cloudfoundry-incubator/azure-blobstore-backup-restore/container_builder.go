package azure

type ContainerBuilder struct {
	config map[string]ContainerConfig
}

func NewContainerBuilder(config map[string]ContainerConfig) ContainerBuilder {
	return ContainerBuilder{config: config}
}

func (c ContainerBuilder) Containers() (map[string]Container, error) {
	var containers = make(map[string]Container)

	for containerId, containerConfig := range c.config {
		container, err := NewSDKContainer(
			containerConfig.Name,
			containerConfig.StorageAccount,
			containerConfig.StorageKey,
			containerConfig.Environment,
		)
		if err != nil {
			return nil, err
		}

		containers[containerId] = container
	}

	return containers, nil
}

func (c ContainerBuilder) RestoreFromStorageAccounts() map[string]StorageAccount {
	var storageAccounts = make(map[string]StorageAccount)

	for containerId, containerConfig := range c.config {
		if containerConfig.RestoreFrom != (RestoreFromConfig{}) {
			storageAccounts[containerId] = StorageAccount{
				Name: containerConfig.RestoreFrom.StorageAccount,
				Key:  containerConfig.RestoreFrom.StorageKey,
			}
		}
	}

	return storageAccounts
}
