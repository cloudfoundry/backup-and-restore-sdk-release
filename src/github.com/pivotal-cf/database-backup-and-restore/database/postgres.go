package database

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
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
	psqlPath, psqlPathVariableSet := os.LookupEnv("PG_CLIENT_PATH")
	if !psqlPathVariableSet {
		log.Fatalln("PG_CLIENT_PATH must be set")
	}

	var pgDumpPath string
	var ok bool
	if ispg94(config, psqlPath) {
		pgDumpPath, ok = os.LookupEnv("PG_DUMP_9_4_PATH")
		if !ok {
			log.Fatalln("PG_DUMP_9_4_PATH must be set")
		}
	} else {
		pgDumpPath, ok = os.LookupEnv("PG_DUMP_9_6_PATH")
		if !ok {
			log.Fatalln("PG_DUMP_9_6_PATH must be set")
		}
	}

	return &postgresBackuper{
		artifactFilePath: artifactFilePath,
		config:           config,
		backupBinary:     pgDumpPath,
	}
}

func NewPostgresRestorer(config Config, artifactFilePath string) *postgresRestorer {
	pgRestorePath, pgRestorePathVariableSet := os.LookupEnv("PG_RESTORE_9_4_PATH")
	if !pgRestorePathVariableSet {
		log.Fatalln("PG_RESTORE_9_4_PATH must be set")
	}
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

	version, _ := ParsePostgresVersion(outb.String())

	return semVer_9_4.MinorVersionMatches(version)
}
