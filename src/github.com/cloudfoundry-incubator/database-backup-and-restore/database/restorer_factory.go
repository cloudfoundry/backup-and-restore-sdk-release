package database

import (
	"github.com/cloudfoundry-incubator/database-backup-and-restore/config"
	"github.com/cloudfoundry-incubator/database-backup-and-restore/mysql"
	"github.com/cloudfoundry-incubator/database-backup-and-restore/postgres"
)

type RestorerFactory struct {
	utilitiesConfig config.UtilitiesConfig
}

func NewRestorerFactory(utilitiesConfig config.UtilitiesConfig) RestorerFactory {
	return RestorerFactory{utilitiesConfig: utilitiesConfig}
}

func (f RestorerFactory) Make(config config.ConnectionConfig) Interactor {
	if config.Adapter == "postgres" {
		return postgres.NewRestorer(
			config, f.utilitiesConfig,
		)
	} else {
		return mysql.NewRestorer(
			config, f.utilitiesConfig,
		)
	}
}
