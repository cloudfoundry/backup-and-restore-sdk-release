package database

import (
	"fmt"
	"log"

	"github.com/cloudfoundry-incubator/database-backup-and-restore/runner"
)

type postgresRestorer struct {
	config        Config
	restoreBinary string
}

type postgresBackuper struct {
	config       Config
	backupBinary string
}

func NewPostgresBackuper(config Config, utilitiesConfig DatabaseUtilitiesConfig) *postgresBackuper {
	psqlPath := utilitiesConfig.Postgres_9_6.Client

	var pgDumpPath string

	if ispg94(config, psqlPath) {
		pgDumpPath = utilitiesConfig.Postgres_9_4.Dump
	} else {
		pgDumpPath = utilitiesConfig.Postgres_9_6.Dump
	}

	return &postgresBackuper{
		config:       config,
		backupBinary: pgDumpPath,
	}
}

func NewPostgresRestorer(config Config, utilitiesConfig DatabaseUtilitiesConfig) *postgresRestorer {
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

func ispg94(config Config, psqlPath string) bool {
	stdout, stderr, err := runner.Run(psqlPath, []string{"--tuples-only",
		fmt.Sprintf("--username=%s", config.Username),
		fmt.Sprintf("--host=%s", config.Host),
		fmt.Sprintf("--port=%d", config.Port),
		config.Database,
		`--command=SELECT VERSION()`},
		map[string]string{"PGPASSWORD": config.Password})

	if err != nil {
		log.Fatalf("Unable to check version of Postgres: %v\n%s\n%s", err, string(stdout), string(stderr))
	}

	version, _ := ParsePostgresVersion(string(stdout)) // TODO: err

	return semVer_9_4.MinorVersionMatches(version)
}
