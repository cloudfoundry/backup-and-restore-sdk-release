package mysql

import (
	"io/ioutil"

	"github.com/cloudfoundry-incubator/database-backup-restore/config"
)

type SSLOptionsProvider interface {
	BuildSSLParams(*config.TlsConfig) []string
}

type LegacySSLOptionsProvider struct{}

func NewLegacySSLOptionsProvider() LegacySSLOptionsProvider {
	return LegacySSLOptionsProvider{}
}

func (LegacySSLOptionsProvider) BuildSSLParams(config *config.TlsConfig) []string {
	var cmdArgs []string
	caCertFile, _ := ioutil.TempFile("", "")
	ioutil.WriteFile(caCertFile.Name(), []byte(config.Cert.Ca), 0777)
	cmdArgs = append(cmdArgs, "--ssl-ca="+caCertFile.Name())
	if !config.SkipHostVerify {
		cmdArgs = append(cmdArgs, "--ssl-verify-server-cert")
	}
	return cmdArgs
}

type DefaultSSLOptionsProvider struct{}

func NewDefaultSSLProvider() DefaultSSLOptionsProvider {
	return DefaultSSLOptionsProvider{}
}

func (DefaultSSLOptionsProvider) BuildSSLParams(config *config.TlsConfig) []string {
	var cmdArgs []string
	caCertFile, _ := ioutil.TempFile("", "")
	ioutil.WriteFile(caCertFile.Name(), []byte(config.Cert.Ca), 0777)
	cmdArgs = append(cmdArgs, "--ssl-ca="+caCertFile.Name())
	if config.SkipHostVerify {
		cmdArgs = append(cmdArgs, "--ssl-mode=VERIFY_CA")
	} else {
		cmdArgs = append(cmdArgs, "--ssl-mode=VERIFY_IDENTITY")
	}
	return cmdArgs
}
