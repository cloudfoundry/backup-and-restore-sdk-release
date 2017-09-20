package postgres

import (
	"fmt"

	"github.com/cloudfoundry-incubator/database-backup-and-restore/config"
	"github.com/cloudfoundry-incubator/database-backup-and-restore/runner"
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
		"--user=" + b.config.Username,
		"--host=" + b.config.Host,
		fmt.Sprintf("--port=%d", b.config.Port),
		"--format=custom",
		"--file=" + artifactFilePath,
		b.config.Database,
	}
	for _, tableName := range b.config.Tables {
		cmdArgs = append(cmdArgs, "-t", tableName)
	}
	_, _, err := runner.Run(
		b.backupBinary,
		cmdArgs,
		map[string]string{"PGPASSWORD": b.config.Password},
	)

	return err
}
