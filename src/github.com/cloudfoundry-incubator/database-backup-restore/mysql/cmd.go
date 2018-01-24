package mysql

import (
	"fmt"

	"io/ioutil"

	"github.com/cloudfoundry-incubator/database-backup-restore/config"
	"github.com/cloudfoundry-incubator/database-backup-restore/runner"
)

func NewMysqlCommand(config config.ConnectionConfig, cmd string) runner.Command {
	cmdArgs := []string{
		"--user=" + config.Username,
		"--host=" + config.Host,
		fmt.Sprintf("--port=%d", config.Port),
	}

	if config.Tls.Cert.Ca != "" {
		caCertFile, _ := ioutil.TempFile("", "")
		ioutil.WriteFile(caCertFile.Name(), []byte(config.Tls.Cert.Ca), 0777)
		cmdArgs = append(cmdArgs, "--ssl-ca="+caCertFile.Name(), "--ssl-mode=VERIFY_IDENTITY")
	}

	return runner.NewCommand(cmd).WithParams(cmdArgs...).WithEnv(map[string]string{"MYSQL_PWD": config.Password})
}
