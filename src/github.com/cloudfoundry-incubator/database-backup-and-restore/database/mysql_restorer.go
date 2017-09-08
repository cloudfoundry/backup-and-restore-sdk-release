package database

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
)

type mysqlRestorer struct {
	config       ConnectionConfig
	clientBinary string
}

func NewMysqlRestorer(config ConnectionConfig, utilitiesConfig UtilitiesConfig) *mysqlRestorer {
	return &mysqlRestorer{
		config:       config,
		clientBinary: utilitiesConfig.Mysql.Restore,
	}
}

func (r mysqlRestorer) Action(artifactFilePath string) error {
	artifactFile, err := os.Open(artifactFilePath)
	checkErr("Error reading from artifact file,", err)

	cmd := exec.Command(r.clientBinary,
		"-v",
		"--user="+r.config.Username,
		"--host="+r.config.Host,
		fmt.Sprintf("--port=%d", r.config.Port),
		r.config.Database,
	)

	cmd.Stdin = bufio.NewReader(artifactFile)
	cmd.Env = append(cmd.Env, "MYSQL_PWD="+r.config.Password)

	return cmd.Run()
}
