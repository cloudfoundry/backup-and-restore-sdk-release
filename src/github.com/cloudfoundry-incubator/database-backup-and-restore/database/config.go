package database

import (
	"log"
	"os"
)

type UtilityPaths struct {
	Client  string
	Dump    string
	Restore string
}

type DatabaseUtilitiesConfig struct {
	Postgres_9_6 UtilityPaths
	Postgres_9_4 UtilityPaths
	Mysql        UtilityPaths
}

func GetDependencies() DatabaseUtilitiesConfig {
	return DatabaseUtilitiesConfig{
		Postgres_9_6: UtilityPaths{
			Client: lookupEnv("PG_CLIENT_PATH"),
			Dump:   lookupEnv("PG_DUMP_9_6_PATH"),
		},
		Postgres_9_4: UtilityPaths{
			Client:  lookupEnv("PG_CLIENT_PATH"),
			Dump:    lookupEnv("PG_DUMP_9_4_PATH"),
			Restore: lookupEnv("PG_RESTORE_9_4_PATH"),
		},
		Mysql: UtilityPaths{
			Client:  lookupEnv("MYSQL_CLIENT_PATH"),
			Dump:    lookupEnv("MYSQL_DUMP_PATH"),
			Restore: lookupEnv("MYSQL_CLIENT_PATH"),
		},
	}
}

func lookupEnv(key string) string {
	psqlPath, psqlPathVariableSet := os.LookupEnv(key)
	if !psqlPathVariableSet {
		log.Fatalln(key + " must be set")
	}
	return psqlPath
}
