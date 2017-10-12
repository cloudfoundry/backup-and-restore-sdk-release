package database

import (
	"fmt"

	"github.com/cloudfoundry-incubator/database-backup-restore/config"
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
	dumpUtilityVersion, _ := i.dumpUtilityVersionDetector.GetVersion()
	serverVersion, _ := i.serverVersionDetector.GetVersion(i.connectionConfig)

	if !serverVersion.MinorVersionMatches(dumpUtilityVersion) {
		return fmt.Errorf("Version mismatch between dump utility %s and the database server %s\n"+
			"the dump utility and the database server must be at the same major and minor version.\n",
			dumpUtilityVersion,
			serverVersion)
	}

	return i.interactor.Action(artifactFilePath)
}
