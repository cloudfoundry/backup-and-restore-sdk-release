package mysql

import (
	"github.com/cloudfoundry-incubator/database-backup-restore/config"
)

type Backuper struct {
	config             config.ConnectionConfig
	backupBinary       string
	useLegacySslFlags  bool
	sslOptionsProvider SSLOptionsProvider
}

func NewBackuper(config config.ConnectionConfig, backupBinary string, sslOptionsProvider SSLOptionsProvider) Backuper {
	return Backuper{
		config:             config,
		backupBinary:       backupBinary,
		sslOptionsProvider: sslOptionsProvider,
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

	_, _, err := NewMysqlCommand(b.config, b.backupBinary, b.sslOptionsProvider).WithParams(cmdArgs...).Run()

	return err
}
