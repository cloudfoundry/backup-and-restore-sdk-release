package database

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
)

type postgresAdapter struct {
}

func (a postgresAdapter) Backup(config Config, artifactFilePath string) *exec.Cmd {
	var cmd *exec.Cmd
	if ispg94(config) {
		cmd = pgdump94(config, artifactFilePath)
	} else {
		cmd = pgdump96(config, artifactFilePath)
	}

	return cmd
}

func ispg94(config Config) bool {
	psqlPath, psqlPathVariableSet := os.LookupEnv("PG_CLIENT_PATH")
	if !psqlPathVariableSet {
		log.Fatalln("PG_CLIENT_PATH must be set")
	}

	var outb, errb bytes.Buffer

	cmd := exec.Command(psqlPath,
		"--tuples-only",
		fmt.Sprintf("--username=%s", config.Username),
		fmt.Sprintf("--host=%s", config.Host),
		fmt.Sprintf("--port=%d", config.Port),
		config.Database,
		`--command=SELECT VERSION()`,
	)
	cmd.Env = append(cmd.Env, "PGPASSWORD="+config.Password)
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err := cmd.Run()
	if err != nil {
		log.Fatalf("Unable to check version of Postgres: %v\n%s", err, errb.String())
	}

	version, _ := PostgresVersionParser(outb.String())

	return semVer_9_4.MinorVersionMatches(version)

}

func pgdump94(config Config, artifactFilePath string) *exec.Cmd {
	pgDump94Path, pgDump94PathVariableSet := os.LookupEnv("PG_DUMP_9_4_PATH")
	if !pgDump94PathVariableSet {
		log.Fatalln("PG_DUMP_9_4_PATH must be set")
	}
	cmd := exec.Command(pgDump94Path,
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

func pgdump96(config Config, artifactFilePath string) *exec.Cmd {
	pgDump96Path, pgDump96PathVariableSet := os.LookupEnv("PG_DUMP_9_6_PATH")
	if !pgDump96PathVariableSet {
		log.Fatalln("PG_DUMP_9_6_PATH must be set")
	}
	cmd := exec.Command(pgDump96Path,
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
	pgRestorePath, pgRestorePathVariableSet := os.LookupEnv("PG_RESTORE_9_6_PATH")

	if !pgRestorePathVariableSet {
		log.Fatalln("PG_RESTORE_9_6_PATH must be set")
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
