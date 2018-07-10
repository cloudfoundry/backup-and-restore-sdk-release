package gcs

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	Name              string `json:"name"`
	ServiceAccountKey string `json:"gcp_service_account_key"`
}

func ParseConfig(configFilePath string) (map[string]Config, error) {
	configContents, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return nil, err
	}

	var config map[string]Config
	err = json.Unmarshal(configContents, &config)
	if err != nil {
		return nil, err
	}

	return config, nil
}
