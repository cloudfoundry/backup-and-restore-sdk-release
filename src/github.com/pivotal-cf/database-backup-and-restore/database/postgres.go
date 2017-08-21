package database

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

type postgresAdapter struct {
}

func (a postgresAdapter) Backup(config Config, artifactFilePath string) *exec.Cmd {
	pgDumpPath, pgDumpPathVariableSet := os.LookupEnv("PG_DUMP_9_4_PATH")

	if !pgDumpPathVariableSet {
		log.Fatalln("PG_DUMP_9_4_PATH must be set")
	}

	cmd := exec.Command(pgDumpPath,
		"-v",
		"--user="+config.Username,
		"--host="+config.Host,
		fmt.Sprintf("--port=%d", config.Port),
		"--format=custom",
		"--file="+artifactFilePath,
		config.Database,
	)
	cmd.Env = append(cmd.Env, "PGPASSWORD="+config.Password)

	return cmd
}

func (a postgresAdapter) Restore(config Config, artifactFilePath string) *exec.Cmd {
	pgRestorePath, pgRestorePathVariableSet := os.LookupEnv("PG_RESTORE_9_4_PATH")

	if !pgRestorePathVariableSet {
		log.Fatalln("PG_RESTORE_9_4_PATH must be set")
	}

	cmd := exec.Command(pgRestorePath,
		"-v",
		"--user="+config.Username,
		"--host="+config.Host,
		fmt.Sprintf("--port=%d", config.Port),
		"--format=custom",
		"--dbname="+config.Database,
		"--clean",
		artifactFilePath,
	)

	cmd.Env = append(cmd.Env, "PGPASSWORD="+config.Password)

	return cmd
}
