package database

import (
	"github.com/cloudfoundry-incubator/database-backup-and-restore/config"
	"github.com/cloudfoundry-incubator/database-backup-and-restore/version"
)

type Interactor interface {
	Action(artifactFilePath string) error
}

type ServerVersionDetector interface {
	GetVersion(config.ConnectionConfig) (version.SemanticVersion, error)
}

type DumpUtilityVersionDetector interface {
	GetVersion() (version.SemanticVersion, error)
}

type Factory interface {
	Make(Action, config.ConnectionConfig) Interactor
}

type Action string
