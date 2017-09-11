package database

import (
	"log"

	"github.com/cloudfoundry-incubator/database-backup-and-restore/config"
	"github.com/cloudfoundry-incubator/database-backup-and-restore/mysql"
	"github.com/cloudfoundry-incubator/database-backup-and-restore/postgres"
	"github.com/cloudfoundry-incubator/database-backup-and-restore/version"
)

type BackuperFactory struct {
	utilitiesConfig                 config.UtilitiesConfig
	postgresServerVersionDetector   ServerVersionDetector
	mysqlServerVersionDetector      ServerVersionDetector
	mysqlDumpUtilityVersionDetector DumpUtilityVersionDetector
}

func NewBackuperFactory(
	utilitiesConfig config.UtilitiesConfig,
	postgresServerVersionDetector ServerVersionDetector,
	mysqlServerVersionDetector ServerVersionDetector,
	mysqlDumpUtilityVersionDetector DumpUtilityVersionDetector) BackuperFactory {
	return BackuperFactory{
		utilitiesConfig:                 utilitiesConfig,
		postgresServerVersionDetector:   postgresServerVersionDetector,
		mysqlServerVersionDetector:      mysqlServerVersionDetector,
		mysqlDumpUtilityVersionDetector: mysqlDumpUtilityVersionDetector,
	}
}

func (f BackuperFactory) Make(config config.ConnectionConfig) Interactor {
	if config.Adapter == "postgres" {
		return f.makePostgresBackuper(config)
	} else {
		return f.makeMysqlBackuper(config)
	}
}

func (f BackuperFactory) makeMysqlBackuper(config config.ConnectionConfig) mysql.Backuper {
	mysqldumpVersion, _ := f.mysqlDumpUtilityVersionDetector.GetVersion()
	log.Printf("%s version %v\n", f.utilitiesConfig.Mysql.Client, mysqldumpVersion)
	serverVersion, _ := f.mysqlServerVersionDetector.GetVersion(config)
	log.Printf("MYSQL server (%s:%d) version %v\n", config.Host, config.Port, serverVersion)
	if !serverVersion.MinorVersionMatches(mysqldumpVersion) {
		log.Fatalf("Version mismatch between mysqldump %s and the MYSQL server %s\n"+
			"mysqldump utility and the MYSQL server must be at the same major and minor version.\n",
			mysqldumpVersion,
			serverVersion)
	}
	return mysql.NewBackuper(
		config, f.utilitiesConfig,
	)
}

func (f BackuperFactory) makePostgresBackuper(config config.ConnectionConfig) Interactor {
	// TODO: err
	postgresVersion, _ := f.postgresServerVersionDetector.GetVersion(config)

	postgres94Version := version.SemanticVersion{Major: "9", Minor: "4"}
	if postgres94Version.MinorVersionMatches(postgresVersion) {
		return postgres.NewBackuper(config, f.utilitiesConfig.Postgres_9_4.Dump)
	} else {
		return postgres.NewBackuper(config, f.utilitiesConfig.Postgres_9_6.Dump)
	}
}
