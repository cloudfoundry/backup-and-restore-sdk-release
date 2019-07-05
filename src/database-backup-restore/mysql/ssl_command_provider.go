package mysql

import (
	"strings"

	"database-backup-restore/config"
)

type SSLOptionsProvider interface {
	BuildSSLParams(*config.TlsConfig) []string
}

type LegacySSLOptionsProvider struct {
	tempFolderManager config.TempFolderManager
}

func NewLegacySSLOptionsProvider(tempFolderManager config.TempFolderManager) LegacySSLOptionsProvider {
	return LegacySSLOptionsProvider{
		tempFolderManager: tempFolderManager,
	}
}

func (p LegacySSLOptionsProvider) BuildSSLParams(config *config.TlsConfig) []string {
	var cmdArgs []string
	if config != nil {
		caFileName, _ := p.tempFolderManager.WriteTempFile(config.Cert.Ca)
		cmdArgs = append(cmdArgs, "--ssl-ca="+caFileName)
		if config.Cert.Certificate != "" {
			certFileName, _ := p.tempFolderManager.WriteTempFile(config.Cert.Certificate)
			cmdArgs = append(cmdArgs, "--ssl-cert="+certFileName)
		}
		if config.Cert.PrivateKey != "" {
			clientKeyFileName, _ := p.tempFolderManager.WriteTempFile(config.Cert.PrivateKey)
			cmdArgs = append(cmdArgs, "--ssl-key="+clientKeyFileName)
		}
		if !config.SkipHostVerify {
			cmdArgs = append(cmdArgs, "--ssl-verify-server-cert")
		}
	} else {
		cmdArgs = append(cmdArgs, "--ssl-cipher="+supportedCipherList())
	}
	return cmdArgs
}

type DefaultSSLOptionsProvider struct {
	tempFolderManager config.TempFolderManager
}

func NewDefaultSSLProvider(tempFolderManager config.TempFolderManager) DefaultSSLOptionsProvider {
	return DefaultSSLOptionsProvider{
		tempFolderManager: tempFolderManager,
	}
}

func (p DefaultSSLOptionsProvider) BuildSSLParams(config *config.TlsConfig) []string {
	var cmdArgs []string
	if config != nil {
		caFileName, _ := p.tempFolderManager.WriteTempFile(config.Cert.Ca)
		cmdArgs = append(cmdArgs, "--ssl-ca="+caFileName)
		if config.Cert.Certificate != "" {
			clientCertFileName, _ := p.tempFolderManager.WriteTempFile(config.Cert.Certificate)
			cmdArgs = append(cmdArgs, "--ssl-cert="+clientCertFileName)
		}
		if config.Cert.PrivateKey != "" {
			clientKeyFileName, _ := p.tempFolderManager.WriteTempFile(config.Cert.PrivateKey)
			cmdArgs = append(cmdArgs, "--ssl-key="+clientKeyFileName)
		}
		if config.SkipHostVerify {
			cmdArgs = append(cmdArgs, "--ssl-mode=VERIFY_CA")
		} else {
			cmdArgs = append(cmdArgs, "--ssl-mode=VERIFY_IDENTITY")
		}
	}
	return cmdArgs
}

func supportedCipherList() string {
	return strings.Join(append(OpenSSLCipherList, YaSSLCipherList...), ":")
}

var OpenSSLCipherList = []string{
	"AES256-GCM-SHA384",
	"AES256-SHA",
	"AES256-SHA256",
	"CAMELLIA256-SHA",
	"DES-CBC3-SHA",
	"DHE-DSS-AES256-GCM-SHA384",
	"DHE-DSS-AES256-SHA",
	"DHE-DSS-AES256-SHA256",
	"DHE-DSS-CAMELLIA256-SHA",
	"DHE-RSA-AES256-GCM-SHA384",
	"DHE-RSA-AES256-SHA",
	"DHE-RSA-AES256-SHA256",
	"DHE-RSA-CAMELLIA256-SHA",
	"ECDH-ECDSA-AES256-GCM-SHA384",
	"ECDH-ECDSA-AES256-SHA",
	"ECDH-ECDSA-AES256-SHA384",
	"ECDH-ECDSA-DES-CBC3-SHA",
	"ECDH-RSA-AES256-GCM-SHA384",
	"ECDH-RSA-AES256-SHA",
	"ECDH-RSA-AES256-SHA384",
	"ECDH-RSA-DES-CBC3-SHA",
	"ECDHE-ECDSA-AES128-GCM-SHA256",
	"ECDHE-ECDSA-AES128-SHA",
	"ECDHE-ECDSA-AES128-SHA256",
	"ECDHE-ECDSA-AES256-GCM-SHA384",
	"ECDHE-ECDSA-AES256-SHA",
	"ECDHE-ECDSA-AES256-SHA384",
	"ECDHE-ECDSA-DES-CBC3-SHA",
	"ECDHE-RSA-AES128-GCM-SHA256",
	"ECDHE-RSA-AES128-SHA",
	"ECDHE-RSA-AES128-SHA256",
	"ECDHE-RSA-AES256-GCM-SHA384",
	"ECDHE-RSA-AES256-SHA",
	"ECDHE-RSA-AES256-SHA384",
	"ECDHE-RSA-DES-CBC3-SHA",
	"EDH-DSS-DES-CBC3-SHA",
	"EDH-RSA-DES-CBC3-SHA",
	"PSK-3DES-EDE-CBC-SHA",
	"PSK-AES256-CBC-SHA",
	"SRP-DSS-3DES-EDE-CBC-SHA",
	"SRP-DSS-AES-128-CBC-SHA",
	"SRP-DSS-AES-256-CBC-SHA",
	"SRP-RSA-3DES-EDE-CBC-SHA",
	"SRP-RSA-AES-128-CBC-S",
	"SRP-RSA-AES-256-CBC-SHA",
}

var YaSSLCipherList = []string{
	"AES128-RMD",
	"AES128-SHA",
	"AES256-RMD",
	"AES256-SHA",
	"DES-CBC-SHA",
	"DES-CBC3-RMD",
	"DES-CBC3-SHA",
	"DHE-RSA-AES128-RMD",
	"DHE-RSA-AES128-SHA",
	"DHE-RSA-AES256-RMD",
	"DHE-RSA-AES256-SHA",
	"DHE-RSA-DES-CBC3-RMD",
	"EDH-RSA-DES-CBC-SHA",
	"EDH-RSA-DES-CBC3-SHA",
	"RC4-MD5",
	"RC4-SHA",
}
