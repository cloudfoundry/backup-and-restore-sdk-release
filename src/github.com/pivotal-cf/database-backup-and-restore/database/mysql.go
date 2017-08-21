package database

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
)

type mysqlAdapter struct {
}

func (a mysqlAdapter) Backup(config Config, artifactFilePath string) *exec.Cmd {
	mysqlDumpPath, mysqlDumpPathVariableSet := os.LookupEnv("MYSQL_DUMP_PATH")
	mysqlClientPath, mysqlClientPathVariableSet := os.LookupEnv("MYSQL_CLIENT_PATH")

	if !mysqlDumpPathVariableSet {
		log.Fatalln("MYSQL_DUMP_PATH must be set")
	}

	if !mysqlClientPathVariableSet {
		log.Fatalln("MYSQL_CLIENT_PATH must be set")
	}

	// sample output: "mysqldump  Ver 10.16 Distrib 10.1.22-MariaDB, for Linux (x86_64)"
	// /mysqldump\s+Ver\s+[^ ]+\s+Distrib\s+([^ ]+),/
	mysqldumpCmd := exec.Command(mysqlDumpPath, "-V")
	mysqldumpVersion := extractVersionUsingCommand(
		mysqldumpCmd,
		`^mysqldump\s+Ver\s+[^ ]+\s+Distrib\s+([^ ]+),`)

	log.Printf("%s version %v\n", mysqlDumpPath, mysqldumpVersion)

	// extract version from mysql server
	mysqlClientCmd := exec.Command(mysqlClientPath,
		"--skip-column-names",
		"--silent",
		fmt.Sprintf("--user=%s", config.Username),
		fmt.Sprintf("--password=%s", config.Password),
		fmt.Sprintf("--host=%s", config.Host),
		fmt.Sprintf("--port=%d", config.Port),
		"--execute=SELECT VERSION()")
	mysqlServerVersion := extractVersionUsingCommand(mysqlClientCmd, `(.+)`)

	log.Printf("MYSQL server (%s:%d) version %v\n", config.Host, config.Port, mysqlServerVersion)

	// compare versions: for ServerX.ServerY.ServerZ and DumpX.DumpY.DumpZ
	// 	=> ServerX != DumpX => error
	//	=> ServerY != DumpY => error
	// ServerZ and DumpZ are regarded as patch version and compatibility is assumed
	if mysqlServerVersion.major != mysqldumpVersion.major || mysqlServerVersion.minor != mysqldumpVersion.minor {
		log.Fatalf("Version mismatch between mysqldump %s and the MYSQL server %s\n"+
			"mysqldump utility and the MYSQL server must be at the same major and minor version.\n",
			mysqldumpVersion,
			mysqlServerVersion)
	}

	cmd := exec.Command(mysqlDumpPath,
		"-v",
		"--single-transaction",
		"--skip-add-locks",
		"--user="+config.Username,
		"--host="+config.Host,
		fmt.Sprintf("--port=%d", config.Port),
		"--result-file="+artifactFilePath,
		config.Database,
	)
	cmd.Env = append(cmd.Env, "MYSQL_PWD="+config.Password)

	return cmd
}

func (a mysqlAdapter) Restore(config Config, artifactFilePath string) *exec.Cmd {
	mysqlClientPath, mysqlClientPathVariableSet := os.LookupEnv("MYSQL_CLIENT_PATH")

	if !mysqlClientPathVariableSet {
		log.Fatalln("MYSQL_CLIENT_PATH must be set")
	}

	artifactFile, err := os.Open(artifactFilePath)
	checkErr("Error reading from artifact file,", err)

	cmd := exec.Command(mysqlClientPath,
		"-v",
		"--user="+config.Username,
		"--host="+config.Host,
		fmt.Sprintf("--port=%d", config.Port),
		config.Database,
	)

	cmd.Stdin = bufio.NewReader(artifactFile)
	cmd.Env = append(cmd.Env, "MYSQL_PWD="+config.Password)

	return cmd
}

func extractVersionUsingCommand(cmd *exec.Cmd, searchPattern string) semanticVersion {
	stdout, err := cmd.Output()
	checkErr("Error running command.", err)

	r := regexp.MustCompile(searchPattern)
	matches := r.FindSubmatch(stdout)
	if matches == nil {
		log.Fatalln("Could not determine version by using search pattern:", searchPattern)
	}

	versionString := matches[1]

	r = regexp.MustCompile(`(\d+).(\d+).(\S+)`)
	matches = r.FindSubmatch(versionString)
	if matches == nil {
		log.Fatalln("Could not determine version by using search pattern:", searchPattern)
	}

	return semanticVersion{
		major: string(matches[1]),
		minor: string(matches[2]),
		patch: string(matches[3]),
	}
}
