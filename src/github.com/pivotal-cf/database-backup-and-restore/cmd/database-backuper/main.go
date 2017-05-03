package main

import (
	"os"
	"os/exec"
)

func main() {
	user := os.Getenv("USER")
	host := os.Getenv("HOST")
	port := os.Getenv("PORT")
	password := os.Getenv("PASSWORD")
	pgDumpPath, pgDumpPathVariableSet := os.LookupEnv("PG_DUMP_PATH")
	if !pgDumpPathVariableSet {
		pgDumpPath = "/var/vcap/packages/database-backuper-postgres-9.4.11/bin/pg_dump"
	}
	outputFile := os.Getenv("OUTPUT_FILE")
	database := os.Getenv("DATABASE")

	cmd := exec.Command(pgDumpPath, "--user="+user, "--host="+host, "--port="+port, "--format=custom", "--file="+outputFile, database)
	cmd.Env = append(cmd.Env, "PGPASSWORD="+password)
	if err := cmd.Run(); err != nil {
		os.Exit(1)
	}
}
