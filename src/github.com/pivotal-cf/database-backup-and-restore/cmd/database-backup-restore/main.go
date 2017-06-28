package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
)

var supportedAdapters = []string{"postgres"}

func isSupported(adapter string) bool {
	for _, el := range supportedAdapters {
		if el == adapter {
			return true
		}
	}
	return false
}

type Config struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	Port       string `json:"port"`
	Adapter    string `json:"adapter"`
	Host       string `json:"host"`
	OutputFile string `json:"output_file"`
	Database   string `json:"database"`
}

func main() {

	var configPath = flag.String("config", "", "Path to JSON config file")
	var backupAction = flag.Bool("backup", false, "Run database backup")
	var restoreAction = flag.Bool("restore", false, "Run database restore")

	flag.Parse()

	if *configPath == "" {
		log.Fatalln("missing argument: --config config.json\nUsage: database-backuper --config config.json")
	}

	if !*backupAction && !*restoreAction {
		log.Fatalln("Missing --backup or --restore flag")
	}

	if *backupAction && *restoreAction {
		log.Fatalln("Only one of: --backup or --restore can be provided")
	}

	configString, err := ioutil.ReadFile(*configPath)
	if err != nil {
		log.Fatalf("Fail reading config file: %s", err)
	}

	var config Config
	if err := json.Unmarshal(configString, &config); err != nil {
		log.Fatalf("Could not parse config json: %s", err)
	}

	if !isSupported(config.Adapter) {
		log.Fatalf("Unsupported adapter %s", config.Adapter)
	}

	var cmd *exec.Cmd
	if *restoreAction {
		cmd = restore(config)
	} else {
		cmd = backup(config)
	}

	fmt.Println(cmd.Args)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func restore(config Config) *exec.Cmd {
	pgRestorePath, pgRestorePathVariableSet := os.LookupEnv("PG_RESTORE_PATH")

	if !pgRestorePathVariableSet {
		log.Fatalln("PG_RESTORE_PATH must be set")
	}

	cmd := exec.Command(pgRestorePath, "-v", "--user="+config.Username, "--host="+config.Host, "--port="+config.Port, "--format=custom", "--dbname="+config.Database, "--clean", config.OutputFile)
	cmd.Env = append(cmd.Env, "PGPASSWORD="+config.Password)

	return cmd
}

func backup(config Config) *exec.Cmd {
	pgDumpPath, pgDumpPathVariableSet := os.LookupEnv("PG_DUMP_PATH")

	if !pgDumpPathVariableSet {
		log.Fatalln("PG_DUMP_PATH must be set")
	}

	cmd := exec.Command(pgDumpPath, "-v", "--user="+config.Username, "--host="+config.Host, "--port="+config.Port, "--format=custom", "--file="+config.OutputFile, config.Database)
	cmd.Env = append(cmd.Env, "PGPASSWORD="+config.Password)

	return cmd
}
