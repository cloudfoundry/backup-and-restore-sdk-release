package database

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/cloudfoundry-incubator/database-backup-and-restore/runner"
)

type postgresRestorer struct {
	artifactFilePath string
	config           Config
	restoreBinary    string
}

type postgresBackuper struct {
	artifactFilePath string
	config           Config
	backupBinary     string
}

func NewPostgresBackuper(config Config, artifactFilePath string) *postgresBackuper {
	psqlPath := lookupEnv("PG_CLIENT_PATH")

	var pgDumpPath string

	if ispg94(config, psqlPath) {
		pgDumpPath = lookupEnv("PG_DUMP_9_4_PATH")

	} else {
		pgDumpPath = lookupEnv("PG_DUMP_9_6_PATH")
	}

	return &postgresBackuper{
		artifactFilePath: artifactFilePath,
		config:           config,
		backupBinary:     pgDumpPath,
	}
}

func NewPostgresRestorer(config Config, artifactFilePath string) *postgresRestorer {
	pgRestorePath := lookupEnv("PG_RESTORE_9_4_PATH")

	return &postgresRestorer{
		artifactFilePath: artifactFilePath,
		config:           config,
		restoreBinary:    pgRestorePath,
	}
}

func (r postgresRestorer) Action() *exec.Cmd {

	cmd := exec.Command(r.restoreBinary,
		"-v",
		"--user="+r.config.Username,
		"--host="+r.config.Host,
		fmt.Sprintf("--port=%d", r.config.Port),
		"--format=custom",
		"--dbname="+r.config.Database,
		"--clean",
		r.artifactFilePath,
	)

	cmd.Env = append(cmd.Env, "PGPASSWORD="+r.config.Password)

	return cmd
}

func (b postgresBackuper) Action() *exec.Cmd {
	cmdArgs := []string{
		"-v",
		"--user=" + b.config.Username,
		"--host=" + b.config.Host,
		fmt.Sprintf("--port=%d", b.config.Port),
		"--format=custom",
		"--file=" + b.artifactFilePath,
		b.config.Database,
	}

	for _, tableName := range b.config.Tables {
		cmdArgs = append(cmdArgs, "-t", tableName)
	}

	cmd := exec.Command(b.backupBinary, cmdArgs...)
	cmd.Env = append(cmd.Env, "PGPASSWORD="+b.config.Password)

	return cmd
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
