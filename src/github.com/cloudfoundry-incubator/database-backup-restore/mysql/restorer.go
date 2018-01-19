package mysql

import (
	"bufio"
	"log"
	"os"

	"github.com/cloudfoundry-incubator/database-backup-restore/config"
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

	_, _, err = NewMysqlCommand(r.config, r.clientBinary).WithParams("-v", r.config.Database).WithStdin(artifactReader).Run()

	return err
}
