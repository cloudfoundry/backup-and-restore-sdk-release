package mysql

import (
	"fmt"
	"log"

	"strings"

	"github.com/cloudfoundry-incubator/database-backup-restore/config"
	"github.com/cloudfoundry-incubator/database-backup-restore/version"
)

type ServerVersionDetector struct {
	mysqlPath string
}

func NewServerVersionDetector(mysqlPath string) ServerVersionDetector {
	return ServerVersionDetector{mysqlPath: mysqlPath}
}

func (d ServerVersionDetector) GetVersion(config config.ConnectionConfig) (version.DatabaseServerVersion, error) {
	stdout, stderr, err := NewMysqlCommand(config, d.mysqlPath).WithParams(
		"--skip-column-names",
		"--silent",
		"--execute=SELECT VERSION()",
	).Run()

	if err != nil {
		return version.DatabaseServerVersion{}, fmt.Errorf(string(stderr))
	}

	versionString := string(stdout)

	semanticVersion, err := version.ParseSemVerFromString(versionString)
	if err != nil {
		return version.DatabaseServerVersion{}, err
	}

	log.Printf("MYSQL server version %v\n", semanticVersion)

	implementation := parseImplementation(versionString)

	return version.DatabaseServerVersion{
		Implementation:  implementation,
		SemanticVersion: semanticVersion,
	}, nil
}

func parseImplementation(versionString string) string {
	if strings.Contains(strings.ToLower(versionString), "mariadb") {
		return "mariadb"
	} else {
		return "mysql"
	}
}
