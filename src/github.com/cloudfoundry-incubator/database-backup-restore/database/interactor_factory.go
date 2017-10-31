package database

import (
	"fmt"

	"github.com/cloudfoundry-incubator/database-backup-restore/config"
	"github.com/cloudfoundry-incubator/database-backup-restore/mysql"
	"github.com/cloudfoundry-incubator/database-backup-restore/postgres"
	"github.com/cloudfoundry-incubator/database-backup-restore/version"
)

type InteractorFactory struct {
	utilitiesConfig               config.UtilitiesConfig
	postgresServerVersionDetector ServerVersionDetector
}

func NewInteractorFactory(
	utilitiesConfig config.UtilitiesConfig,
	postgresServerVersionDetector ServerVersionDetector) InteractorFactory {

	return InteractorFactory{
		utilitiesConfig:               utilitiesConfig,
		postgresServerVersionDetector: postgresServerVersionDetector,
	}
}

func (f InteractorFactory) Make(action Action, config config.ConnectionConfig) (Interactor, error) {
	switch {
	case config.Adapter == "postgres" && action == "backup":
		return f.makePostgresBackuper(config)
	case config.Adapter == "mysql" && action == "backup":
		return f.makeMysqlBackuper(config), nil
	case config.Adapter == "postgres" && action == "restore":
		return f.makePostgresRestorer(config)
	case config.Adapter == "mysql" && action == "restore":
		return mysql.NewRestorer(config, f.utilitiesConfig.Mysql.Restore), nil
	}

	return nil, fmt.Errorf("unsupported adapter/action combination: %s/%s", config.Adapter, action)
}

func (f InteractorFactory) makeMysqlBackuper(config config.ConnectionConfig) Interactor {
	return NewVersionSafeInteractor(
		mysql.NewBackuper(config, f.utilitiesConfig.Mysql.Dump),
		mysql.NewServerVersionDetector(f.utilitiesConfig.Mysql.Client),
		mysql.NewMysqlDumpUtilityVersionDetector(f.utilitiesConfig.Mysql.Dump),
		config)
}

func (f InteractorFactory) makePostgresBackuper(config config.ConnectionConfig) (Interactor, error) {
	postgresVersion, err := f.postgresServerVersionDetector.GetVersion(config)
	if err != nil {
		return nil, err
	}

	psqlPath, pgDumpPath, _, err := f.getUtilitiesForPostgres(postgresVersion)
	if err != nil {
		return nil, err
	}

	postgresBackuper := postgres.NewBackuper(config, pgDumpPath)
	tableChecker := postgres.NewTableChecker(config, psqlPath)
	return NewTableCheckingInteractor(config, tableChecker, postgresBackuper), nil
}

func (f InteractorFactory) makePostgresRestorer(config config.ConnectionConfig) (Interactor, error) {
	postgresVersion, err := f.postgresServerVersionDetector.GetVersion(config)
	if err != nil {
		return nil, err
	}

	_, _, pgRestorePath, err := f.getUtilitiesForPostgres(postgresVersion)
	if err != nil {
		return nil, err
	}

	return postgres.NewRestorer(config, pgRestorePath), nil
}

func (f InteractorFactory) getUtilitiesForPostgres(postgresVersion version.SemanticVersion) (string, string, string, error) {
	var psqlPath, pgDumpPath, pgRestorePath string
	if version.V_9_4.MinorVersionMatches(postgresVersion) {
		psqlPath = f.utilitiesConfig.Postgres94.Client
		pgDumpPath = f.utilitiesConfig.Postgres94.Dump
		pgRestorePath = f.utilitiesConfig.Postgres94.Restore
	} else if version.V_9_6.MinorVersionMatches(postgresVersion) {
		psqlPath = f.utilitiesConfig.Postgres96.Client
		pgDumpPath = f.utilitiesConfig.Postgres96.Dump
		pgRestorePath = f.utilitiesConfig.Postgres96.Restore
	} else {
		return "", "", "", fmt.Errorf("unsupported version of postgresql: %s.%s", postgresVersion.Major, postgresVersion.Minor)
	}

	return psqlPath, pgDumpPath, pgRestorePath, nil
}
