package database

import (
	"errors"
	"fmt"
	"os"

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
		if mysqlVersion.SemanticVersion.MinorVersionMatches(version.SemVer("5", "6", "38")) {
			return f.utilitiesConfig.Mysql56.Dump, f.utilitiesConfig.Mysql56.Restore, nil
		}
		if mysqlVersion.SemanticVersion.MinorVersionMatches(version.SemVer("5", "7", "20")) {
			return f.utilitiesConfig.Mysql57.Dump, f.utilitiesConfig.Mysql57.Restore, nil
		}
		if mysqlVersion.SemanticVersion.MinorVersionMatches(version.SemVer("8", "0", "0")) {
			if _, err := os.Stat(f.utilitiesConfig.Mysql80.Client) ; os.IsNotExist(err) {
				return "", "", errors.New("MySQL 8.0 is not supported on this OS. Are you using an old (xenial?) stemcell?")
			}
			return f.utilitiesConfig.Mysql80.Dump, f.utilitiesConfig.Mysql80.Restore, nil
		}
	}

	return "", "", fmt.Errorf("unsupported version of %s: %s.%s", implementation, semVer.Major, semVer.Minor)
}

func (f InteractorFactory) getSSLCommandProvider(mysqlVersion version.DatabaseServerVersion) mysql.SSLOptionsProvider {
	switch {
	case mysqlVersion.SemanticVersion.MinorVersionMatches(version.SemVer("5", "7", "20")),
		mysqlVersion.SemanticVersion.MinorVersionMatches(version.SemVer("8", "0", "0")):
		return mysql.NewDefaultSSLProvider(f.tempFolderManager)
	default:
		return mysql.NewLegacySSLOptionsProvider(f.tempFolderManager)
	}
}

func (f InteractorFactory) getAdditionalOptionsProvider(mysqlVersion version.DatabaseServerVersion) mysql.AdditionalOptionsProvider {
	if mysqlVersion.Implementation == "mariadb" || mysqlVersion.SemanticVersion.MinorVersionMatches(version.SemVer("5", "5", "20")) {
		return mysql.NewEmptyAdditionalOptionsProvider()
	} else {
		return mysql.NewPurgeGTIDOptionProvider()
	}
}

func (f InteractorFactory) getUtilitiesForPostgres(postgresVersion version.DatabaseServerVersion) (string, string, string, error) {
	semVer := postgresVersion.SemanticVersion
	if semVer.MinorVersionMatches(version.SemVer("9", "4", "x")) {
		return f.utilitiesConfig.Postgres94.Client,
			f.utilitiesConfig.Postgres94.Dump,
			f.utilitiesConfig.Postgres94.Restore,
			nil
	} else if semVer.MinorVersionMatches(version.SemVer("9", "6", "x")) {
		return f.utilitiesConfig.Postgres96.Client,
			f.utilitiesConfig.Postgres96.Dump,
			f.utilitiesConfig.Postgres96.Restore,
			nil
	} else if semVer.MajorVersionMatches(version.SemVer("10", "x", "x")) {
		return f.utilitiesConfig.Postgres10.Client,
			f.utilitiesConfig.Postgres10.Dump,
			f.utilitiesConfig.Postgres10.Restore,
			nil
	} else if semVer.MajorVersionMatches(version.SemVer("11", "x", "x")) {
		return f.utilitiesConfig.Postgres11.Client,
			f.utilitiesConfig.Postgres11.Dump,
			f.utilitiesConfig.Postgres11.Restore,
			nil
	} else if semVer.MajorVersionMatches(version.SemVer("13", "x", "x")) {
		return f.utilitiesConfig.Postgres13.Client,
			f.utilitiesConfig.Postgres13.Dump,
			f.utilitiesConfig.Postgres13.Restore,
			nil
	}

	return "", "", "", fmt.Errorf("unsupported version of postgresql: %s.%s", semVer.Major, semVer.Minor)
}
