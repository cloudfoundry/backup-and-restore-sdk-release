package database

import (
	"fmt"

	"github.com/cloudfoundry-incubator/database-backup-and-restore/runner"
)

type postgresBackuper struct {
	config       ConnectionConfig
	backupBinary string
}

func NewPostgresBackuper(config ConnectionConfig, backupBinary string) *postgresBackuper {
	return &postgresBackuper{
		config:       config,
		backupBinary: backupBinary,
	}
}

func (b postgresBackuper) Action(artifactFilePath string) error {
	cmdArgs := []string{
		"-v",
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
