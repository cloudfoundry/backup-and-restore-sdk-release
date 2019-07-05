package azure

import "fmt"

type Environment string

const (
	DefaultEnvironment      Environment = "AzureCloud"
	ChinaEnvironment        Environment = "AzureChinaCloud"
	USGovernmentEnvironment Environment = "AzureUSGovernment"
	GermanEnvironment       Environment = "AzureGermanCloud"
)

func (e Environment) Suffix() (string, error) {
	suffixes := map[Environment]string{
		DefaultEnvironment:      "core.windows.net",
		ChinaEnvironment:        "core.chinacloudapi.cn",
		USGovernmentEnvironment: "core.usgovcloudapi.net",
		GermanEnvironment:       "core.cloudapi.de",
	}

	suffix, ok := suffixes[e]
	if !ok {
		return "", fmt.Errorf("invalid environment: %s", e)
	}

	return suffix, nil
}
