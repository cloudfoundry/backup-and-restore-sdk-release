package config

import (
	"log"
	"os"
)

type UtilityPaths struct {
	Client  string
	Dump    string
	Restore string
}

type UtilitiesConfig struct {
	Postgres96 UtilityPaths
	Postgres94 UtilityPaths
	Mysql      UtilityPaths
}

func GetDependencies() UtilitiesConfig {
	return UtilitiesConfig{
		Postgres96: UtilityPaths{
			Client: lookupEnv("PG_CLIENT_PATH"),
			Dump:   lookupEnv("PG_DUMP_9_6_PATH"),
		},
		Postgres94: UtilityPaths{
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

var supportedAdapters = []string{"postgres", "mysql"}

func isSupported(adapter string) bool {
	for _, el := range supportedAdapters {
		if el == adapter {
			return true
		}
	}
	return false
}
