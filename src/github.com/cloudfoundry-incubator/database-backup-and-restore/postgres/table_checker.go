package postgres

import (
	"fmt"

	"strings"

	"github.com/cloudfoundry-incubator/database-backup-and-restore/config"
	"github.com/cloudfoundry-incubator/database-backup-and-restore/runner"
)

type TableChecker struct {
	config   config.ConnectionConfig
	psqlPath string
}

func NewTableChecker(config config.ConnectionConfig, psqlPath string) TableChecker {
	return TableChecker{config: config, psqlPath: psqlPath}
}

func (c TableChecker) FindMissingTables(tableNames []string) ([]string, error) {
	stdout, _, err := runner.Run(c.psqlPath,
		[]string{"--tuples-only",
			fmt.Sprintf("--username=%s", c.config.Username),
			fmt.Sprintf("--host=%s", c.config.Host),
			fmt.Sprintf("--port=%d", c.config.Port),
			c.config.Database,
			`--command=SELECT table_name FROM information_schema.tables WHERE table_type='BASE TABLE' AND table_schema='public';`,
		}, map[string]string{"PGPASSWORD": c.config.Password})
	if err != nil {
		return nil, err
	}

	databaseTables := NewTableSet(strings.Split(string(stdout), "\n"))

	missingTables := []string{}
	for _, tableName := range tableNames {
		if !databaseTables.Contains(tableName) {
			missingTables = append(missingTables, tableName)
		}
	}

	return missingTables, nil
}

type TableSet []string

func NewTableSet(tables []string) TableSet {
	return tables
}

func (s TableSet) Contains(str string) bool {
	for _, element := range s {
		if element == str {
			return true
		}
	}
	return false
}
