package database

import (
	"fmt"

	"github.com/cloudfoundry-incubator/database-backup-and-restore/runner"
)

type postgresRestorer struct {
	config        ConnectionConfig
	restoreBinary string
}

func NewPostgresRestorer(config ConnectionConfig, utilitiesConfig UtilitiesConfig) *postgresRestorer {
	return &postgresRestorer{
		config:        config,
		restoreBinary: utilitiesConfig.Postgres_9_4.Restore,
	}
}

func (r postgresRestorer) Action(artifactFilePath string) error {
	_, _, err := runner.Run(r.restoreBinary, []string{"-v",
		"--user=" + r.config.Username,
		"--host=" + r.config.Host,
		fmt.Sprintf("--port=%d", r.config.Port),
		"--format=custom",
		"--dbname=" + r.config.Database,
		"--clean",
		artifactFilePath},
		map[string]string{"PGPASSWORD": r.config.Password})

	return err
}
