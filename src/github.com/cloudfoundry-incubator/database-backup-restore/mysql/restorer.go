package mysql

import (
	"bufio"
	"log"
	"os"

	"github.com/cloudfoundry-incubator/database-backup-restore/config"
)

type Restorer struct {
	config             config.ConnectionConfig
	clientBinary       string
	sslOptionsProvider SSLOptionsProvider
}

func NewRestorer(
	config config.ConnectionConfig,
	restoreBinary string,
	sslOptionsProvider SSLOptionsProvider) Restorer {
	return Restorer{
		config:             config,
		clientBinary:       restoreBinary,
		sslOptionsProvider: sslOptionsProvider,
	}
}

func (r Restorer) Action(artifactFilePath string) error {
	artifactFile, err := os.Open(artifactFilePath)
	if err != nil {
		log.Fatalln("Error reading from artifact file,", err)
	}
	artifactReader := bufio.NewReader(artifactFile)

	_, _, err = NewMysqlCommand(
		r.config,
		r.clientBinary,
		r.sslOptionsProvider).WithParams("-v", r.config.Database).WithStdin(artifactReader).Run()

	return err
}
