package mysql

import (
	"github.com/cloudfoundry-incubator/backup-and-restore-sdk/database-backup-restore/config"
)

type Backuper struct {
	config                    config.ConnectionConfig
	backupBinary              string
	useLegacySslFlags         bool
	sslOptionsProvider        SSLOptionsProvider
	additionalOptionsProvider AdditionalOptionsProvider
}

func NewBackuper(config config.ConnectionConfig, backupBinary string, sslOptionsProvider SSLOptionsProvider, additionalOptionsProvider AdditionalOptionsProvider) Backuper {
	return Backuper{
		config:                    config,
		backupBinary:              backupBinary,
		sslOptionsProvider:        sslOptionsProvider,
		additionalOptionsProvider: additionalOptionsProvider,
	}
}

func (b Backuper) Action(artifactFilePath string) error {
	cmdArgs := []string{
		"-v",
		"--skip-add-locks",
		"--single-transaction",
		"--result-file=" + artifactFilePath,
		b.config.Database,
	}

	cmdArgs = append(cmdArgs, b.config.Tables...)
	cmdArgs = append(cmdArgs, b.additionalOptionsProvider.BuildParams()...)

	_, _, err := NewMysqlCommand(b.config, b.backupBinary, b.sslOptionsProvider).WithParams(cmdArgs...).Run()

	return err
}
