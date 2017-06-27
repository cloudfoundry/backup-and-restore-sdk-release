package system_tests

import (
	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"fmt"
	"time"
)

var _ = Describe("restore", func() {
	var databaseName string
	var dbDumpPath string
	var configPath string

	BeforeEach(func() {
		databaseName = "db" + strconv.FormatInt(time.Now().Unix(), 10)
		runOnPostgresVM(fmt.Sprintf(`/var/vcap/packages/postgres-9.4/bin/createdb -U vcap "%s"`, databaseName))
		runSqlCommand("CREATE TABLE people (name varchar);", databaseName)
		runSqlCommand("INSERT into people VALUES ('Derik');", databaseName)

		configPath = "/tmp/config.json" + strconv.FormatInt(time.Now().Unix(), 10)
		dbDumpPath = "/tmp/sql_dump" + strconv.FormatInt(time.Now().Unix(), 10)

		configJson := fmt.Sprintf(
			`{"username":"vcap","password":"%s","host":"localhost","port":"5432","database":"%s","adapter":"postgres","output_file":"%s"}`,
			MustHaveEnv("POSTGRES_PASSWORD"),
			databaseName,
			dbDumpPath,
		)
		runOnPostgresVM(fmt.Sprintf("echo '%s' > %s", configJson, configPath))
	})

	AfterEach(func() {
		runOnPostgresVM(fmt.Sprintf(`/var/vcap/packages/postgres-9.4/bin/dropdb -U vcap "%s"`, databaseName))
		runOnPostgresVM(fmt.Sprintf("rm -rf %s", configPath))
		runOnPostgresVM(fmt.Sprintf("rm -rf %s", dbDumpPath))
	})

	Context("database-backuper is colocated with Postgres", func() {
		It("restores the Postgres database", func() {
			runOnPostgresVM(fmt.Sprintf(`/var/vcap/jobs/database-backuper/bin/backup %s`, configPath))

			runSqlCommand("UPDATE people SET NAME = 'Dave';", databaseName)

			runOnPostgresVM(fmt.Sprintf("/var/vcap/jobs/database-backuper/bin/restore %s", configPath))

			Expect(runSqlCommand("SELECT name FROM people;", databaseName)).To(gbytes.Say("Derik"))
			Expect(runSqlCommand("SELECT name FROM people;", databaseName)).NotTo(gbytes.Say("Dave"))
		})
	})
})
