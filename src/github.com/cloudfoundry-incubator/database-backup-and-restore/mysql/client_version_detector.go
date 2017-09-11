package mysql

import (
	"os/exec"

	"github.com/cloudfoundry-incubator/database-backup-and-restore/version"
)

type ClientVersionDetector struct {
	mysqldumpPath string
}

func NewMysqlDumpUtilityVersionDetector(mysqldumpPath string) ClientVersionDetector {
	return ClientVersionDetector{mysqldumpPath: mysqldumpPath}
}

func (d ClientVersionDetector) GetVersion() (version.SemanticVersion, error) {
	// sample output: "mysqldump  Ver 10.16 Distrib 10.1.22-MariaDB, for Linux (x86_64)"
	// /mysqldump\s+Ver\s+[^ ]+\s+Distrib\s+([^ ]+),/
	clientCmd := exec.Command(d.mysqldumpPath, "-V")

	return extractVersionUsingCommand(clientCmd, `^mysqldump\s+Ver\s+[^ ]+\s+Distrib\s+([^ ]+),`), nil
}
