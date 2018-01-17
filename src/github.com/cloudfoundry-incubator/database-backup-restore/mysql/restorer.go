package mysql

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/cloudfoundry-incubator/database-backup-restore/config"
	"github.com/cloudfoundry-incubator/database-backup-restore/runner"
)

type Restorer struct {
	config       config.ConnectionConfig
	clientBinary string
}

func NewRestorer(config config.ConnectionConfig, restoreBinary string) Restorer {
	return Restorer{
		config:       config,
		clientBinary: restoreBinary,
	}
}

func (r Restorer) Action(artifactFilePath string) error {
	artifactFile, err := os.Open(artifactFilePath)
	if err != nil {
		log.Fatalln("Error reading from artifact file,", err)
	}
	artifactReader := bufio.NewReader(artifactFile)

	_, _, err = runner.RunWithStdin(r.clientBinary, []string{
		"-v",
		"--user=" + r.config.Username,
		"--host=" + r.config.Host,
		fmt.Sprintf("--port=%d", r.config.Port),
		r.config.Database},
		map[string]string{"MYSQL_PWD": r.config.Password},
		artifactReader)

	return err
}
