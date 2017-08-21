package database

import (
	"log"
	"os/exec"
)

type Adapter interface {
	Backup(config Config, artifactFilePath string) *exec.Cmd
	Restore(config Config, artifactFilePath string) *exec.Cmd
}

type Config struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Port     int    `json:"port"`
	Adapter  string `json:"adapter"`
	Host     string `json:"host"`
	Database string `json:"database"`
}

func GetAdapter(adapter string) Adapter {
	if adapter == "postgres" {
		return postgresAdapter{}
	} else {
		return mysqlAdapter{}
	}
}

func checkErr(msg string, err error) {
	if err != nil {
		log.Fatalln(msg, err)
	}
}
