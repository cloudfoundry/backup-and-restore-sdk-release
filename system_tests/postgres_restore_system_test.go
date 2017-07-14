package system_tests

import (
	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"fmt"
	"time"
)

var _ = Describe("postgres-restore", func() {
	var databaseName string
	var dbDumpPath string
	var configPath string

	BeforeEach(func() {
		databaseName = "db" + strconv.FormatInt(time.Now().Unix(), 10)
		runOnPostgresVMAndSucceed(fmt.Sprintf(`/var/vcap/packages/postgres-9.4/bin/createdb -U vcap "%s"`, databaseName))
		runPostgresSqlCommand("CREATE TABLE people (name varchar);", databaseName)
		runPostgresSqlCommand("INSERT INTO people VALUES ('Derik');", databaseName)

		configPath = "/tmp/config.json" + strconv.FormatInt(time.Now().Unix(), 10)
		dbDumpPath = "/tmp/sql_dump" + strconv.FormatInt(time.Now().Unix(), 10)

		configJson := fmt.Sprintf(
			`{"username":"vcap","password":"%s","host":"localhost","port":5432,"database":"%s","adapter":"postgres"}`,
			MustHaveEnv("POSTGRES_PASSWORD"),
			databaseName,
		)
		runOnPostgresVMAndSucceed(fmt.Sprintf("echo '%s' > %s", configJson, configPath))
	})

	AfterEach(func() {
		runOnPostgresVMAndSucceed(fmt.Sprintf(`/var/vcap/packages/postgres-9.4/bin/dropdb -U vcap "%s"`, databaseName))
		runOnPostgresVMAndSucceed(fmt.Sprintf("rm -rf %s %s", configPath, dbDumpPath))
	})

	Context("database-backup-restorer is colocated with Postgres", func() {
		It("restores the Postgres database", func() {
			runOnPostgresVMAndSucceed(fmt.Sprintf(`/var/vcap/jobs/database-backup-restorer/bin/backup --config %s --artifact-file %s`, configPath, dbDumpPath))

			runPostgresSqlCommand("UPDATE people SET NAME = 'Dave';", databaseName)

			runOnPostgresVMAndSucceed(fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/restore --config %s --artifact-file %s", configPath, dbDumpPath))

			Expect(runPostgresSqlCommand("SELECT name FROM people;", databaseName)).To(gbytes.Say("Derik"))
			Expect(runPostgresSqlCommand("SELECT name FROM people;", databaseName)).NotTo(gbytes.Say("Dave"))
		})
	})
})
