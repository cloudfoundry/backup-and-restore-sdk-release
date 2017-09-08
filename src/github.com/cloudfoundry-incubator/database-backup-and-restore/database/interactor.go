package database

import (
	"github.com/cloudfoundry-incubator/database-backup-and-restore/config"
	"github.com/cloudfoundry-incubator/database-backup-and-restore/version"
)

type Interactor interface {
	Action(artifactFilePath string) error
}

type VersionDetector interface {
	GetVersion(config.ConnectionConfig) (version.SemanticVersion, error)
}

type Factory interface {
	Make(config.ConnectionConfig) Interactor
}
