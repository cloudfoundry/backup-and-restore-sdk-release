package database

import (
	"log"

	"github.com/cloudfoundry-incubator/database-backup-and-restore/config"
	"github.com/cloudfoundry-incubator/database-backup-and-restore/mysql"
)

type ComplexMysqlInteractor struct {
	mysql.Backuper
	serverVersionDetector      ServerVersionDetector
	dumpUtilityVersionDetector DumpUtilityVersionDetector
	connectionConfig           config.ConnectionConfig
}

func NewCompoundMysqlInteractor(
	backuper mysql.Backuper,
	serverVersionDetector ServerVersionDetector,
	dumpUtilityVersionDetector DumpUtilityVersionDetector,
	config config.ConnectionConfig,
) ComplexMysqlInteractor {
	return ComplexMysqlInteractor{
		Backuper:                   backuper,
		serverVersionDetector:      serverVersionDetector,
		dumpUtilityVersionDetector: dumpUtilityVersionDetector,
		connectionConfig:           config,
	}
}
func (i ComplexMysqlInteractor) Action(artifactFilePath string) error {
	mysqldumpVersion, _ := i.dumpUtilityVersionDetector.GetVersion()
	log.Printf("Mysql dump version %v\n", mysqldumpVersion)
	serverVersion, _ := i.serverVersionDetector.GetVersion(i.connectionConfig)
	log.Printf("MYSQL server version %v\n", serverVersion)
	if !serverVersion.MinorVersionMatches(mysqldumpVersion) {
		log.Fatalf("Version mismatch between mysqldump %s and the MYSQL server %s\n"+
			"mysqldump utility and the MYSQL server must be at the same major and minor version.\n",
			mysqldumpVersion,
			serverVersion)
	}
	return i.Backuper.Action(artifactFilePath)
}
