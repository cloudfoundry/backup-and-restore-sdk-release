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
	mysqlServerVersionDetector    ServerVersionDetector
	tempFolderManager             config.TempFolderManager
}

func NewInteractorFactory(
	utilitiesConfig config.UtilitiesConfig,
	postgresServerVersionDetector ServerVersionDetector,
	mysqlServerVersionDetector ServerVersionDetector,
	tempFolderManager config.TempFolderManager) InteractorFactory {

	return InteractorFactory{
		utilitiesConfig:               utilitiesConfig,
		postgresServerVersionDetector: postgresServerVersionDetector,
		mysqlServerVersionDetector:    mysqlServerVersionDetector,
		tempFolderManager:             tempFolderManager,
	}
}

func (f InteractorFactory) Make(action Action, connectionConfig config.ConnectionConfig) (Interactor, error) {
	switch {
	case connectionConfig.Adapter == "postgres" && action == "backup":
		return f.makePostgresBackuper(connectionConfig)
	case connectionConfig.Adapter == "mysql" && action == "backup":
		return f.makeMysqlBackuper(connectionConfig)
	case connectionConfig.Adapter == "postgres" && action == "restore":
		return f.makePostgresRestorer(connectionConfig)
	case connectionConfig.Adapter == "mysql" && action == "restore":
		return f.makeMysqlRestorer(connectionConfig)
	}

	return nil, fmt.Errorf("unsupported adapter/action combination: %s/%s", connectionConfig.Adapter, action)
}

func (f InteractorFactory) makeMysqlBackuper(config config.ConnectionConfig) (Interactor, error) {
	mysqldbVersion, err := f.mysqlServerVersionDetector.GetVersion(config, f.tempFolderManager)
	if err != nil {
		return nil, err
	}

	mysqlDumpPath, _, err := f.getUtilitiesForMySQL(mysqldbVersion)
	if err != nil {
		return nil, err
	}

	mysqlSSLProvider := f.getSSLCommandProvider(mysqldbVersion)
	mysqlAdditionalOptionsProvider := f.getAdditionalOptionsProvider(mysqldbVersion)

	return mysql.NewBackuper(config, mysqlDumpPath, mysqlSSLProvider, mysqlAdditionalOptionsProvider), nil
}

func (f InteractorFactory) makeMysqlRestorer(config config.ConnectionConfig) (Interactor, error) {
	mysqldbVersion, err := f.mysqlServerVersionDetector.GetVersion(config, f.tempFolderManager)
	if err != nil {
		return nil, err
	}

	_, mysqlRestorePath, err := f.getUtilitiesForMySQL(mysqldbVersion)
	if err != nil {
		return nil, err
	}

	mysqlSSLProvider := f.getSSLCommandProvider(mysqldbVersion)

	return mysql.NewRestorer(config, mysqlRestorePath, mysqlSSLProvider), nil
}

func (f InteractorFactory) makePostgresBackuper(config config.ConnectionConfig) (Interactor, error) {
	postgresVersion, err := f.postgresServerVersionDetector.GetVersion(config, f.tempFolderManager)
	if err != nil {
		return nil, err
	}

	psqlPath, pgDumpPath, _, err := f.getUtilitiesForPostgres(postgresVersion)
	if err != nil {
		return nil, err
	}

	postgresBackuper := postgres.NewBackuper(config, f.tempFolderManager, pgDumpPath)
	tableChecker := postgres.NewTableChecker(config, psqlPath)
	return NewTableCheckingInteractor(config, tableChecker, postgresBackuper), nil
}

func (f InteractorFactory) makePostgresRestorer(config config.ConnectionConfig) (Interactor, error) {
	postgresVersion, err := f.postgresServerVersionDetector.GetVersion(config, f.tempFolderManager)
	if err != nil {
		return nil, err
	}

	_, _, pgRestorePath, err := f.getUtilitiesForPostgres(postgresVersion)
	if err != nil {
		return nil, err
	}

	return postgres.NewRestorer(config, f.tempFolderManager, pgRestorePath), nil
}

func (f InteractorFactory) getUtilitiesForMySQL(mysqlVersion version.DatabaseServerVersion) (string, string, error) {
	implementation := mysqlVersion.Implementation
	semVer := mysqlVersion.SemanticVersion
	switch {
	case implementation == "mariadb" && semVer.MinorVersionMatches(version.SemVer("10", "1", "24")):
		return f.utilitiesConfig.Mariadb.Dump, f.utilitiesConfig.Mariadb.Restore, nil
	case implementation == "mysql":
		if mysqlVersion.SemanticVersion.MinorVersionMatches(version.SemVer("5", "5", "58")) {
			return f.utilitiesConfig.Mysql55.Dump, f.utilitiesConfig.Mysql55.Restore, nil
		}
		if mysqlVersion.SemanticVersion.MinorVersionMatches(version.SemVer("5", "6", "38")) {
			return f.utilitiesConfig.Mysql56.Dump, f.utilitiesConfig.Mysql56.Restore, nil
		}
		if mysqlVersion.SemanticVersion.MinorVersionMatches(version.SemVer("5", "7", "20")) {
			return f.utilitiesConfig.Mysql57.Dump, f.utilitiesConfig.Mysql57.Restore, nil
		}
	}

	return "", "", fmt.Errorf("unsupported version of %s: %s.%s", implementation, semVer.Major, semVer.Minor)
}

func (f InteractorFactory) getSSLCommandProvider(mysqlVersion version.DatabaseServerVersion) mysql.SSLOptionsProvider {
	if mysqlVersion.SemanticVersion.MinorVersionMatches(version.SemVer("5", "7", "20")) {
		return mysql.NewDefaultSSLProvider(f.tempFolderManager)
	} else {
		return mysql.NewLegacySSLOptionsProvider(f.tempFolderManager)
	}
}

func (f InteractorFactory) getAdditionalOptionsProvider(mysqlVersion version.DatabaseServerVersion) mysql.AdditionalOptionsProvider {
	if mysqlVersion.SemanticVersion.MinorVersionMatches(version.SemVer("5", "5", "20")) {
		return mysql.NewLegacyAdditionalOptionsProvider()
	} else {
		return mysql.NewDefaultAdditionalOptionsProvider()
	}
}

func (f InteractorFactory) getUtilitiesForPostgres(postgresVersion version.DatabaseServerVersion) (string, string, string, error) {
	semVer := postgresVersion.SemanticVersion
	if semVer.MinorVersionMatches(version.SemVer("9", "4", "11")) {
		return f.utilitiesConfig.Postgres94.Client,
			f.utilitiesConfig.Postgres94.Dump,
			f.utilitiesConfig.Postgres94.Restore,
			nil
	} else if semVer.MinorVersionMatches(version.SemVer("9", "6", "3")) {
		return f.utilitiesConfig.Postgres96.Client,
			f.utilitiesConfig.Postgres96.Dump,
			f.utilitiesConfig.Postgres96.Restore,
			nil
	}

	return "", "", "", fmt.Errorf("unsupported version of postgresql: %s.%s", semVer.Major, semVer.Minor)
}
