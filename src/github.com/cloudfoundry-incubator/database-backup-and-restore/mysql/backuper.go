package mysql

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
		"-v",
		"--single-transaction",
		"--skip-add-locks",
		"--user=" + b.config.Username,
		"--host=" + b.config.Host,
		fmt.Sprintf("--port=%d", b.config.Port),
		"--result-file=" + artifactFilePath,
		b.config.Database,
	}

	cmdArgs = append(cmdArgs, b.config.Tables...)

	_, _, err := runner.Run(b.backupBinary, cmdArgs, map[string]string{"MYSQL_PWD": b.config.Password})

	return err
}
