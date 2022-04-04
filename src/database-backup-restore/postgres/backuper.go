package postgres

import (
	"github.com/cloudfoundry/backup-and-restore-sdk-release/src/database-backup-restore/config"
)

type Backuper struct {
	config            config.ConnectionConfig
	tempFolderManager config.TempFolderManager
	backupBinary      string
}

func NewBackuper(config config.ConnectionConfig, tempFolderManager config.TempFolderManager, backupBinary string) Backuper {
	return Backuper{
		config:            config,
		tempFolderManager: tempFolderManager,
		backupBinary:      backupBinary,
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

	_, _, err := NewPostgresCommand(b.config, b.tempFolderManager, b.backupBinary).WithParams(cmdArgs...).Run()

	return err
}
