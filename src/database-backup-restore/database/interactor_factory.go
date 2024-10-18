package database

import (
	"fmt"

	"database-backup-restore/config"
	"database-backup-restore/mysql"
	"database-backup-restore/postgres"
	"database-backup-restore/version"
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
	case implementation == "mariadb" && semVer.MajorVersionMatches(version.SemVer("10", "x", "x")):
		return f.utilitiesConfig.Mariadb.Dump, f.utilitiesConfig.Mariadb.Restore, nil
	case implementation == "mysql":
		if mysqlVersion.SemanticVersion.MinorVersionMatches(version.SemVer("5", "7", "0")) {
			return f.utilitiesConfig.Mysql57.Dump, f.utilitiesConfig.Mysql57.Restore, nil
		}
		if mysqlVersion.SemanticVersion.MajorVersionMatches(version.SemVer("8", "0", "0")) {
			return f.utilitiesConfig.Mysql80.Dump, f.utilitiesConfig.Mysql80.Restore, nil
		}
	}

	return "", "", fmt.Errorf("unsupported version of %s: %s.%s", implementation, semVer.Major, semVer.Minor)
}

func (f InteractorFactory) getSSLCommandProvider(mysqlVersion version.DatabaseServerVersion) mysql.SSLOptionsProvider {
	switch {
	case mysqlVersion.SemanticVersion.MajorVersionMatches(version.SemVer("5", "7", "0")), mysqlVersion.SemanticVersion.MajorVersionMatches(version.SemVer("8", "0", "0")):
		return mysql.NewDefaultSSLProvider(f.tempFolderManager)
	default:
		return mysql.NewLegacySSLOptionsProvider(f.tempFolderManager)
	}
}

func (f InteractorFactory) getAdditionalOptionsProvider(mysqlVersion version.DatabaseServerVersion) mysql.AdditionalOptionsProvider {
	if mysqlVersion.Implementation != "mariadb" {
		return mysql.NewPurgeGTIDOptionProvider()
	}

	return mysql.NewEmptyAdditionalOptionsProvider()
}

func (f InteractorFactory) getUtilitiesForPostgres(postgresVersion version.DatabaseServerVersion) (string, string, string, error) {
	semVer := postgresVersion.SemanticVersion
	if semVer.MajorVersionMatches(version.SemVer("13", "x", "x")) {
		return f.utilitiesConfig.Postgres13.Client,
			f.utilitiesConfig.Postgres13.Dump,
			f.utilitiesConfig.Postgres13.Restore,
			nil
	} else if semVer.MajorVersionMatches(version.SemVer("15", "x", "x")) {
		return f.utilitiesConfig.Postgres15.Client,
			f.utilitiesConfig.Postgres15.Dump,
			f.utilitiesConfig.Postgres15.Restore,
			nil
	}

	return "", "", "", fmt.Errorf("unsupported version of postgresql: %s.%s", semVer.Major, semVer.Minor)
}
