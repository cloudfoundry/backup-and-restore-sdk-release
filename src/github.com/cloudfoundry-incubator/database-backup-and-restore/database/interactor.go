package database

import (
	"github.com/cloudfoundry-incubator/database-backup-and-restore/config"
	"github.com/cloudfoundry-incubator/database-backup-and-restore/version"
)

//go:generate counterfeiter -o fakes/fake_interactor.go . Interactor
type Interactor interface {
	Action(artifactFilePath string) error
}

//go:generate counterfeiter -o fakes/fake_server_version_detector.go . ServerVersionDetector
type ServerVersionDetector interface {
	GetVersion(config.ConnectionConfig) (version.SemanticVersion, error)
}

//go:generate counterfeiter -o fakes/fake_dump_utility_version_detector.go . DumpUtilityVersionDetector
type DumpUtilityVersionDetector interface {
	GetVersion() (version.SemanticVersion, error)
}

type Factory interface {
	Make(Action, config.ConnectionConfig) Interactor
}

type Action string
