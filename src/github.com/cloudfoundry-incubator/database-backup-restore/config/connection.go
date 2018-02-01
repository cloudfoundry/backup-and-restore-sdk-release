package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type ConnectionConfig struct {
	Username string    `json:"username"`
	Password string    `json:"password"`
	Port     int       `json:"port"`
	Adapter  string    `json:"adapter"`
	Host     string    `json:"host"`
	Database string    `json:"database"`
	Tables   []string  `json:"tables"`
	Tls      TlsConfig `json:"tls"`
}

type TlsConfig struct {
	SkipHostVerify bool          `json:skip_host_verify`
	Cert           CertTlsConfig `json:"cert"`
}

type CertTlsConfig struct {
	Ca          string `json:"ca"`
	Certificate string `json:"certificate"`
	PrivateKey  string `json:"private_key"`
}

func ParseAndValidateConnectionConfig(configPath string) (ConnectionConfig, error) {
	configString, err := ioutil.ReadFile(configPath)
	if err != nil {
		return ConnectionConfig{}, fmt.Errorf("Fail reading config file: %s\n", err)
	}

	var connectionConfig ConnectionConfig
	if err := json.Unmarshal(configString, &connectionConfig); err != nil {
		return ConnectionConfig{}, fmt.Errorf("Could not parse config json: %s\n", err)
	}

	if !isSupported(connectionConfig.Adapter) {
		return ConnectionConfig{}, fmt.Errorf("Unsupported adapter %s\n", connectionConfig.Adapter)
	}

	if connectionConfig.Tables != nil && len(connectionConfig.Tables) == 0 {
		return ConnectionConfig{}, fmt.Errorf("Tables specified but empty\n")
	}

	return connectionConfig, nil
}

var supportedAdapters = []string{"postgres", "mysql"}

func isSupported(adapter string) bool {
	for _, el := range supportedAdapters {
		if el == adapter {
			return true
		}
	}
	return false
}
