package azure

import (
	"encoding/json"
	"io/ioutil"
)

type ContainerConfig struct {
	Name           string `json:"name"`
	StorageAccount string `json:"azure_storage_account"`
	StorageKey     string `json:"azure_storage_key"`
}

func ParseConfig(configFilePath string) (map[string]ContainerConfig, error) {
	configContents, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return nil, err
	}

	var config map[string]ContainerConfig
	err = json.Unmarshal(configContents, &config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
