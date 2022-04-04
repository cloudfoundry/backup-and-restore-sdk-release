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
	Postgres10 UtilityPaths
	Postgres11 UtilityPaths
	Postgres13 UtilityPaths
	Mariadb    UtilityPaths
	Mysql56    UtilityPaths
	Mysql57    UtilityPaths
}

func GetUtilitiesConfigFromEnv() UtilitiesConfig {
	return UtilitiesConfig{
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
		Postgres10: UtilityPaths{
			Client:  lookupEnv("PG_CLIENT_PATH"),
			Dump:    lookupEnv("PG_DUMP_10_PATH"),
			Restore: lookupEnv("PG_RESTORE_10_PATH"),
		},
		Postgres96: UtilityPaths{
			Client:  lookupEnv("PG_CLIENT_PATH"),
			Dump:    lookupEnv("PG_DUMP_9_6_PATH"),
			Restore: lookupEnv("PG_RESTORE_9_6_PATH"),
		},
		Postgres94: UtilityPaths{
			Client:  lookupEnv("PG_CLIENT_PATH"),
			Dump:    lookupEnv("PG_DUMP_9_4_PATH"),
			Restore: lookupEnv("PG_RESTORE_9_4_PATH"),
		},
		Mariadb: UtilityPaths{
			Client:  lookupEnv("MARIADB_CLIENT_PATH"),
			Dump:    lookupEnv("MARIADB_DUMP_PATH"),
			Restore: lookupEnv("MARIADB_CLIENT_PATH"),
		},
		Mysql56: UtilityPaths{
			Client:  lookupEnv("MYSQL_CLIENT_5_6_PATH"),
			Dump:    lookupEnv("MYSQL_DUMP_5_6_PATH"),
			Restore: lookupEnv("MYSQL_CLIENT_5_6_PATH"),
		},
		Mysql57: UtilityPaths{
			Client:  lookupEnv("MYSQL_CLIENT_5_7_PATH"),
			Dump:    lookupEnv("MYSQL_DUMP_5_7_PATH"),
			Restore: lookupEnv("MYSQL_CLIENT_5_7_PATH"),
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
