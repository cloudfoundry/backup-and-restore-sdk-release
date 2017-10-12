package postgres

import (
	"fmt"

	"strings"

	"github.com/cloudfoundry-incubator/database-backup-restore/config"
	"github.com/cloudfoundry-incubator/database-backup-restore/runner"
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

	tableList := parseTableList(string(stdout))
	databaseTables := NewTableSet(tableList)

	missingTables := []string{}
	for _, tableName := range tableNames {
		if !databaseTables.Contains(tableName) {
			missingTables = append(missingTables, tableName)
		}
	}

	return missingTables, nil
}

func parseTableList(tableColumn string) []string {
	untrimmedTables := strings.Split(tableColumn, "\n")

	var trimmedTables []string
	for _, untrimmedTable := range untrimmedTables {
		trimmedTables = append(trimmedTables, strings.TrimSpace(untrimmedTable))
	}
	return trimmedTables
}

type TableSet map[string]bool

func NewTableSet(tables []string) TableSet {
	tableSet := map[string]bool{}

	for _, table := range tables {
		tableSet[table] = true
	}

	return tableSet
}

func (s TableSet) Contains(str string) bool {
	return s[str]
}
