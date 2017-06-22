package main

import (
	"encoding/json"
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
	pgDumpPath, pgDumpPathVariableSet := os.LookupEnv("PG_DUMP_PATH")
	if !pgDumpPathVariableSet {
		pgDumpPath = "/var/vcap/packages/database-backuper-postgres-9.4.11/bin/pg_dump"
	}

	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "missing argument: config.json\nUsage: database-backuper config.json")
		os.Exit(1)
	}
	configPath := os.Args[1]
	configString, err := ioutil.ReadFile(configPath)
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

	cmd := exec.Command(pgDumpPath, "--user="+config.Username, "--host="+config.Host, "--port="+config.Port, "--format=custom", "--file="+config.OutputFile, config.Database)
	cmd.Env = append(cmd.Env, "PGPASSWORD="+config.Password)
	if err := cmd.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
