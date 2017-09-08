package database

type BackuperFactory struct {
	utilitiesConfig         UtilitiesConfig
	postgresVersionDetector VersionDetector
}

type RestorerFactory struct {
	utilitiesConfig UtilitiesConfig
}

type VersionDetector interface {
	GetVersion(ConnectionConfig) (semanticVersion, error)
}

type Factory interface {
	Make(ConnectionConfig) DBInteractor
}

func NewBackuperFactory(utilitiesConfig UtilitiesConfig, postgresVersionDetector VersionDetector) BackuperFactory {
	return BackuperFactory{
		utilitiesConfig:         utilitiesConfig,
		postgresVersionDetector: postgresVersionDetector,
	}
}

func (f BackuperFactory) Make(config ConnectionConfig) DBInteractor {
	if config.Adapter == "postgres" {
		return f.makePostgresBackuper(config)
	} else {
		return NewMysqlBackuper(
			config, f.utilitiesConfig,
		)
	}
}

func (f BackuperFactory) makePostgresBackuper(config ConnectionConfig) DBInteractor {
	var pgDumpPath string

	// TODO: err
	postgresVersion, _ := f.postgresVersionDetector.GetVersion(config)

	postgres94Version := semanticVersion{major: "9", minor: "4"}
	if postgres94Version.MinorVersionMatches(postgresVersion) {
		pgDumpPath = f.utilitiesConfig.Postgres_9_4.Dump
	} else {
		pgDumpPath = f.utilitiesConfig.Postgres_9_6.Dump
	}

	return NewPostgresBackuper(config, pgDumpPath)
}

func NewRestorerFactory(utilitiesConfig UtilitiesConfig) RestorerFactory {
	return RestorerFactory{utilitiesConfig: utilitiesConfig}
}

func (f RestorerFactory) Make(config ConnectionConfig) DBInteractor {
	if config.Adapter == "postgres" {
		return NewPostgresRestorer(
			config, f.utilitiesConfig,
		)
	} else {
		return NewMysqlRestorer(
			config, f.utilitiesConfig,
		)
	}
}
