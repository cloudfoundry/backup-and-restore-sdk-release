package database

import (
	"log"

	"github.com/cloudfoundry-incubator/database-backup-and-restore/config"
)

type VersionSafeInteractor struct {
	interactor                 Interactor
	serverVersionDetector      ServerVersionDetector
	dumpUtilityVersionDetector DumpUtilityVersionDetector
	connectionConfig           config.ConnectionConfig
}

func NewVersionSafeInteractor(
	interactor Interactor,
	serverVersionDetector ServerVersionDetector,
	dumpUtilityVersionDetector DumpUtilityVersionDetector,
	config config.ConnectionConfig,
) VersionSafeInteractor {
	return VersionSafeInteractor{
		interactor:                 interactor,
		serverVersionDetector:      serverVersionDetector,
		dumpUtilityVersionDetector: dumpUtilityVersionDetector,
		connectionConfig:           config,
	}
}

func (i VersionSafeInteractor) Action(artifactFilePath string) error {
	mysqldumpVersion, _ := i.dumpUtilityVersionDetector.GetVersion()
	serverVersion, _ := i.serverVersionDetector.GetVersion(i.connectionConfig)

	if !serverVersion.MinorVersionMatches(mysqldumpVersion) {
		log.Fatalf("Version mismatch between mysqldump %s and the MYSQL server %s\n"+
			"mysqldump utility and the MYSQL server must be at the same major and minor version.\n",
			mysqldumpVersion,
			serverVersion)
	}

	return i.interactor.Action(artifactFilePath)
}
