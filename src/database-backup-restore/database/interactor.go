package database

import (
	"database-backup-restore/config"
	"database-backup-restore/version"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate -o fakes/fake_interactor.go . Interactor
type Interactor interface {
	Action(artifactFilePath string) error
}

//counterfeiter:generate -o fakes/fake_server_version_detector.go . ServerVersionDetector
type ServerVersionDetector interface {
	GetVersion(config.ConnectionConfig, config.TempFolderManager) (version.DatabaseServerVersion, error)
}

//counterfeiter:generate -o fakes/fake_dump_utility_version_detector.go . DumpUtilityVersionDetector
type DumpUtilityVersionDetector interface {
	GetVersion() (version.SemanticVersion, error)
}

type Factory interface {
	Make(Action, config.ConnectionConfig) Interactor
}

type Action string
