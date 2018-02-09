package postgres

import (
	"fmt"

	"io/ioutil"

	"github.com/cloudfoundry-incubator/database-backup-restore/config"
	"github.com/cloudfoundry-incubator/database-backup-restore/runner"
)

type Backuper struct {
	config       config.ConnectionConfig
	backupBinary string
}

func NewBackuper(config config.ConnectionConfig, backupBinary string) Backuper {
	return Backuper{
		config:       config,
		backupBinary: backupBinary,
	}
}

func (b Backuper) Action(artifactFilePath string) error {
	cmdArgs := []string{
		"--verbose",
		"--user=" + b.config.Username,
		"--host=" + b.config.Host,
		fmt.Sprintf("--port=%d", b.config.Port),
		"--format=custom",
		"--file=" + artifactFilePath,
		b.config.Database,
	}

	for _, tableName := range b.config.Tables {
		cmdArgs = append(cmdArgs, "-t", tableName)
	}

	caCertFile, err := ioutil.TempFile("", "")
	if err != nil {
		return err
	}

	env := map[string]string{
		"PGPASSWORD": b.config.Password,
	}

	if b.config.Tls != nil {
		ioutil.WriteFile(caCertFile.Name(), []byte(b.config.Tls.Cert.Ca), 0777)
		env["PGSSLROOTCERT"] = caCertFile.Name()
		env["PGSSLMODE"] = "verify-ca"
	}

	_, _, err = runner.NewCommand(b.backupBinary).
		WithParams(cmdArgs...).
		WithEnv(env).
		Run()

	return err
}
