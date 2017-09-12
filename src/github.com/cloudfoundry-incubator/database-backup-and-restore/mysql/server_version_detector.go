package mysql

import (
	"fmt"
	"log"

	"os/exec"

	"github.com/cloudfoundry-incubator/database-backup-and-restore/config"
	"github.com/cloudfoundry-incubator/database-backup-and-restore/version"
)

type ServerVersionDetector struct {
	mysqlPath string
}

func NewServerVersionDetector(mysqlPath string) ServerVersionDetector {
	return ServerVersionDetector{mysqlPath: mysqlPath}
}

func (d ServerVersionDetector) GetVersion(config config.ConnectionConfig) (version.SemanticVersion, error) {
	clientCmd := exec.Command(d.mysqlPath,
		"--skip-column-names",
		"--silent",
		fmt.Sprintf("--user=%s", config.Username),
		fmt.Sprintf("--password=%s", config.Password),
		fmt.Sprintf("--host=%s", config.Host),
		fmt.Sprintf("--port=%d", config.Port),
		"--execute=SELECT VERSION()")

	semanticVersion := extractVersionUsingCommand(clientCmd, `(.+)`)

	log.Printf("MYSQL server version %v\n", semanticVersion)

	return semanticVersion, nil
}
