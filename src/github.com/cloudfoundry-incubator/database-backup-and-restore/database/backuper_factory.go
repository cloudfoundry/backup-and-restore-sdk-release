package database

import (
	"github.com/cloudfoundry-incubator/database-backup-and-restore/config"
	"github.com/cloudfoundry-incubator/database-backup-and-restore/mysql"
	"github.com/cloudfoundry-incubator/database-backup-and-restore/postgres"
	"github.com/cloudfoundry-incubator/database-backup-and-restore/version"
)

type BackuperFactory struct {
	utilitiesConfig         config.UtilitiesConfig
	postgresVersionDetector VersionDetector
}

func NewBackuperFactory(utilitiesConfig config.UtilitiesConfig, postgresVersionDetector VersionDetector) BackuperFactory {
	return BackuperFactory{
		utilitiesConfig:         utilitiesConfig,
		postgresVersionDetector: postgresVersionDetector,
	}
}

func (f BackuperFactory) Make(config config.ConnectionConfig) Interactor {
	if config.Adapter == "postgres" {
		return f.makePostgresBackuper(config)
	} else {
		return mysql.NewBackuper(
			config, f.utilitiesConfig,
		)
	}
}

func (f BackuperFactory) makePostgresBackuper(config config.ConnectionConfig) Interactor {
	var pgDumpPath string

	// TODO: err
	postgresVersion, _ := f.postgresVersionDetector.GetVersion(config)

	postgres94Version := version.SemanticVersion{Major: "9", Minor: "4"}
	if postgres94Version.MinorVersionMatches(postgresVersion) {
		pgDumpPath = f.utilitiesConfig.Postgres_9_4.Dump
	} else {
		pgDumpPath = f.utilitiesConfig.Postgres_9_6.Dump
	}

	return postgres.NewBackuper(config, pgDumpPath)
}
