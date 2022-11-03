package azure

import (
	"encoding/json"
)

type ContainerConfig struct {
	Name           string            `json:"name"`
	StorageAccount string            `json:"azure_storage_account"`
	StorageKey     string            `json:"azure_storage_key"`
	Environment    Environment       `json:"environment"`
	RestoreFrom    RestoreFromConfig `json:"restore_from"`
}

type RestoreFromConfig struct {
	StorageAccount string `json:"azure_storage_account"`
	StorageKey     string `json:"azure_storage_key"`
}

func ParseConfig(configFilePath string) (map[string]ContainerConfig, error) {
	configContents, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, err
	}

	var config map[string]ContainerConfig
	err = json.Unmarshal(configContents, &config)
	if err != nil {
		return nil, err
	}

	for key, containerConfig := range config {
		if config[key].Environment == "" {
			containerConfig.Environment = "AzureCloud"
			config[key] = containerConfig
		}
	}

	return config, nil
}
