package database

import (
	"fmt"
	"log"

	"github.com/cloudfoundry-incubator/database-backup-and-restore/runner"
)

type PostgresVersionDetector struct {
	psqlPath string
}

func NewPostgresVersionDetector(psqlPath string) PostgresVersionDetector {
	return PostgresVersionDetector{psqlPath: psqlPath}
}

func (d PostgresVersionDetector) GetVersion(config ConnectionConfig) (semanticVersion, error) {
	stdout, stderr, err := runner.Run(d.psqlPath, []string{"--tuples-only",
		fmt.Sprintf("--username=%s", config.Username),
		fmt.Sprintf("--host=%s", config.Host),
		fmt.Sprintf("--port=%d", config.Port),
		config.Database,
		`--command=SELECT VERSION()`},
		map[string]string{"PGPASSWORD": config.Password})

	if err != nil {
		log.Fatalf("Unable to check version of Postgres: %v\n%s\n%s", err, string(stdout), string(stderr))
	}

	return ParsePostgresVersion(string(stdout))
}
