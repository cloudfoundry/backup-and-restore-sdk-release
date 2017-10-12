package database

import (
	"fmt"

	"strings"

	"github.com/cloudfoundry-incubator/database-backup-restore/config"
)

//go:generate counterfeiter -o fakes/fake_table_checker.go . TableChecker
type TableChecker interface {
	FindMissingTables(tableNames []string) ([]string, error)
}

type TableCheckingInteractor struct {
	config       config.ConnectionConfig
	tableChecker TableChecker
	interactor   Interactor
}

func NewTableCheckingInteractor(
	config config.ConnectionConfig,
	tableChecker TableChecker,
	interactor Interactor) TableCheckingInteractor {

	return TableCheckingInteractor{
		config:       config,
		tableChecker: tableChecker,
		interactor:   interactor,
	}
}

func (i TableCheckingInteractor) Action(artifactFilePath string) error {
	if i.config.Tables != nil {
		missingTables, err := i.tableChecker.FindMissingTables(i.config.Tables)
		if err != nil {
			return err
		}
		if len(missingTables) != 0 {
			return fmt.Errorf("can't find specified table(s): %s",
				strings.Join(missingTables, ", "))
		}
	}
	return i.interactor.Action(artifactFilePath)
}
