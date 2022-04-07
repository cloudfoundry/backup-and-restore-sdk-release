package database

import (
	"github.com/cloudfoundry/backup-and-restore-sdk-release/src/database-backup-restore/config"
	"github.com/cloudfoundry/backup-and-restore-sdk-release/src/database-backup-restore/version"
)

//go:generate counterfeiter -o fakes/fake_interactor.go . Interactor
type Interactor interface {
	Action(artifactFilePath string) error
}

//go:generate counterfeiter -o fakes/fake_server_version_detector.go . ServerVersionDetector
type ServerVersionDetector interface {
	GetVersion(config.ConnectionConfig, config.TempFolderManager) (version.DatabaseServerVersion, error)
}

//go:generate counterfeiter -o fakes/fake_dump_utility_version_detector.go . DumpUtilityVersionDetector
type DumpUtilityVersionDetector interface {
	GetVersion() (version.SemanticVersion, error)
}

type Factory interface {
	Make(Action, config.ConnectionConfig) Interactor
}

type Action string
