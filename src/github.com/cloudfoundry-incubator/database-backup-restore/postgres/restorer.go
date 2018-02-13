package postgres

import (
	"fmt"
	"os"

	"io/ioutil"

	"github.com/cloudfoundry-incubator/database-backup-restore/config"
	"github.com/cloudfoundry-incubator/database-backup-restore/runner"
)

type Restorer struct {
	config        config.ConnectionConfig
	restoreBinary string
}

func NewRestorer(config config.ConnectionConfig, restoreBinary string) Restorer {
	return Restorer{
		config:        config,
		restoreBinary: restoreBinary,
	}
}

func (r Restorer) Action(artifactFilePath string) error {
	stdout, _, err := runner.NewCommand(r.restoreBinary).WithParams("--list", artifactFilePath).Run()

	if err != nil {
		return err
	}

	listFile, err := ioutil.TempFile("", "backup-restore-sdk")
	if err != nil {
		return err
	}
	defer os.Remove(listFile.Name())

	listFile.Write(ListFileFilter(stdout))

	cmdArgs := []string{
		"--verbose",
		"--format=custom",
		"--dbname=" + r.config.Database,
		"--clean",
		fmt.Sprintf("--use-list=%s", listFile.Name()),
		artifactFilePath,
	}

	_, _, err = NewPostgresCommand(r.config, r.restoreBinary).WithParams(cmdArgs...).Run()

	return err
}
