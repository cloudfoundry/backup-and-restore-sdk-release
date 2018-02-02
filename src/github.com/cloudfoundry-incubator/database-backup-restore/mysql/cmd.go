package mysql

import (
	"fmt"

	"github.com/cloudfoundry-incubator/database-backup-restore/config"
	"github.com/cloudfoundry-incubator/database-backup-restore/runner"
)

func NewMysqlCommand(config config.ConnectionConfig, cmd string, sslOptionsProvider SSLOptionsProvider) runner.Command {
	cmdArgs := []string{
		"--user=" + config.Username,
		"--host=" + config.Host,
		fmt.Sprintf("--port=%d", config.Port),
	}

	if config.Tls != nil {
		cmdArgs = append(cmdArgs, sslOptionsProvider.BuildSSLParams(config.Tls)...)
	}

	return runner.NewCommand(cmd).WithParams(cmdArgs...).WithEnv(map[string]string{"MYSQL_PWD": config.Password})
}
