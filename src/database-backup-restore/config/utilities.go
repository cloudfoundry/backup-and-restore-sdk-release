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
	Postgres11 UtilityPaths
	Postgres13 UtilityPaths
	Postgres15 UtilityPaths
	Mariadb    UtilityPaths
	Mysql80    UtilityPaths
}

func GetUtilitiesConfigFromEnv() UtilitiesConfig {
	return UtilitiesConfig{
		Postgres15: UtilityPaths{
			Client:  lookupEnv("PG_CLIENT_PATH"),
			Dump:    lookupEnv("PG_DUMP_15_PATH"),
			Restore: lookupEnv("PG_RESTORE_15_PATH"),
		},
		Postgres13: UtilityPaths{
			Client:  lookupEnv("PG_CLIENT_PATH"),
			Dump:    lookupEnv("PG_DUMP_13_PATH"),
			Restore: lookupEnv("PG_RESTORE_13_PATH"),
		},
		Postgres11: UtilityPaths{
			Client:  lookupEnv("PG_CLIENT_PATH"),
			Dump:    lookupEnv("PG_DUMP_11_PATH"),
			Restore: lookupEnv("PG_RESTORE_11_PATH"),
		},
		Mariadb: UtilityPaths{
			Client:  lookupEnv("MARIADB_CLIENT_PATH"),
			Dump:    lookupEnv("MARIADB_DUMP_PATH"),
			Restore: lookupEnv("MARIADB_CLIENT_PATH"),
		},
		Mysql80: UtilityPaths{
			Client:  lookupEnv("MYSQL_CLIENT_8_0_PATH"),
			Dump:    lookupEnv("MYSQL_DUMP_8_0_PATH"),
			Restore: lookupEnv("MYSQL_CLIENT_8_0_PATH"),
		},
	}
}

func lookupEnv(key string) string {
	value, valueSet := os.LookupEnv(key)
	if !valueSet {
		log.Fatalln(key + " must be set")
	}
	return value
}
