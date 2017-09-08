package database

import (
	"log"
)

type Config struct {
	Username string   `json:"username"`
	Password string   `json:"password"`
	Port     int      `json:"port"`
	Adapter  string   `json:"adapter"`
	Host     string   `json:"host"`
	Database string   `json:"database"`
	Tables   []string `json:"tables"`
}

func checkErr(msg string, err error) {
	if err != nil {
		log.Fatalln(msg, err)
	}
}

type DBInteractor interface {
	Action(artifactFilePath string) error
}
