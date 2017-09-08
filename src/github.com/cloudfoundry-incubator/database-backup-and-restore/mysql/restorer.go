package mysql

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/cloudfoundry-incubator/database-backup-and-restore/config"
)

type Restorer struct {
	config       config.ConnectionConfig
	clientBinary string
}

func NewRestorer(config config.ConnectionConfig, utilitiesConfig config.UtilitiesConfig) Restorer {
	return Restorer{
		config:       config,
		clientBinary: utilitiesConfig.Mysql.Restore,
	}
}

func (r Restorer) Action(artifactFilePath string) error {
	artifactFile, err := os.Open(artifactFilePath)
	if err != nil {
		log.Fatalln("Error reading from artifact file,", err)
	}

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
