// Copyright (C) 2017-Present Pivotal Software, Inc. All rights reserved.
//
// This program and the accompanying materials are made available under
// the terms of the under the Apache License, Version 2.0 (the "License”);
// you may not use this file except in compliance with the License.
//
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

var supportedAdapters = []string{"postgres", "mysql"}

func isSupported(adapter string) bool {
	for _, el := range supportedAdapters {
		if el == adapter {
			return true
		}
	}
	return false
}

type Config struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Port     int    `json:"port"`
	Adapter  string `json:"adapter"`
	Host     string `json:"host"`
	Database string `json:"database"`
}

func main() {
	var configPath = flag.String("config", "", "Path to JSON config file")
	var backupAction = flag.Bool("backup", false, "Run database backup")
	var restoreAction = flag.Bool("restore", false, "Run database restore")
	var artifactFilePath = flag.String("artifact-file", "", "Path to output file")

	flag.Parse()

	if *backupAction && *restoreAction {
		failAndPrintUsage("Only one of: --backup or --restore can be provided")
	}

	if *configPath == "" {
		failAndPrintUsage("Missing --config flag")
	}

	if !*backupAction && !*restoreAction {
		failAndPrintUsage("Missing --backup or --restore flag")
	}

	if *artifactFilePath == "" {
		failAndPrintUsage("Missing --artifact-file flag")
	}

	configString, err := ioutil.ReadFile(*configPath)
	if err != nil {
		log.Fatalf("Fail reading config file: %s\n", err)
	}

	var config Config
	if err := json.Unmarshal(configString, &config); err != nil {
		log.Fatalf("Could not parse config json: %s\n", err)
	}

	if !isSupported(config.Adapter) {
		log.Fatalf("Unsupported adapter %s\n", config.Adapter)
	}

	var cmd *exec.Cmd
	if *restoreAction {
		cmd = restore(config, *artifactFilePath)
	} else {
		cmd = backup(config, *artifactFilePath)
	}

	fmt.Println(cmd.Args)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func failAndPrintUsage(message string) {
	log.Fatalf("%s\nUsage: database-backup-restorer [--backup|--restore] --config <config-file> --artifact-file <artifact-file>\n", message)
}

func restore(config Config, artifactFilePath string) *exec.Cmd {
	if config.Adapter == "postgres" {
		return pgRestore(config, artifactFilePath)
	} else {
		return mysqlRestore(config, artifactFilePath)
	}
}

func backup(config Config, artifactFilePath string) *exec.Cmd {
	if config.Adapter == "postgres" {
		return pgDump(config, artifactFilePath)
	} else {
		return mysqlDump(config, artifactFilePath)
	}
}

func pgRestore(config Config, artifactFilePath string) *exec.Cmd {
	pgRestorePath, pgRestorePathVariableSet := os.LookupEnv("PG_RESTORE_PATH")

	if !pgRestorePathVariableSet {
		log.Fatalln("PG_RESTORE_PATH must be set")
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

func mysqlRestore(config Config, artifactFilePath string) *exec.Cmd {
	mysqlRestorePath, mysqlRestorePathVariableSet := os.LookupEnv("MYSQL_CLIENT_PATH")

	if !mysqlRestorePathVariableSet {
		log.Fatalln("MYSQL_CLIENT_PATH must be set")
	}

	artifactFile, err := os.Open(artifactFilePath)
	checkErr("Error reading from artifact file,", err)

	cmd := exec.Command(mysqlRestorePath,
		"-v",
		"--user="+config.Username,
		"--host="+config.Host,
		fmt.Sprintf("--port=%d", config.Port),
		config.Database,
	)

	fmt.Println(ioutil.ReadFile(artifactFilePath))

	cmd.Stdin = bufio.NewReader(artifactFile)
	cmd.Env = append(cmd.Env, "MYSQL_PWD="+config.Password)

	return cmd
}

func pgDump(config Config, artifactFilePath string) *exec.Cmd {
	pgDumpPath, pgDumpPathVariableSet := os.LookupEnv("PG_DUMP_PATH")

	if !pgDumpPathVariableSet {
		log.Fatalln("PG_DUMP_PATH must be set")
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

func mysqlDump(config Config, artifactFilePath string) *exec.Cmd {
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

func checkErr(msg string, err error) {
	if err != nil {
		log.Fatalln(msg, err)
	}
}

type semanticVersion struct {
	major string
	minor string
	patch string
}

func (v semanticVersion) String() string {
	return strings.Join([]string{v.major, v.minor, v.patch}, ".")
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
