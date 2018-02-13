package postgres

import (
	"github.com/cloudfoundry-incubator/database-backup-restore/config"
)

type Backuper struct {
	config       config.ConnectionConfig
	backupBinary string
}

func NewBackuper(config config.ConnectionConfig, backupBinary string) Backuper {
	return Backuper{
		config:       config,
		backupBinary: backupBinary,
	}
}

func (b Backuper) Action(artifactFilePath string) error {
	cmdArgs := []string{
		"--verbose",
		"--format=custom",
		"--file=" + artifactFilePath,
		b.config.Database,
	}

	for _, tableName := range b.config.Tables {
		cmdArgs = append(cmdArgs, "-t", tableName)
	}

	_, _, err := NewPostgresCommand(b.config, b.backupBinary).WithParams(cmdArgs...).Run()

	return err
}
