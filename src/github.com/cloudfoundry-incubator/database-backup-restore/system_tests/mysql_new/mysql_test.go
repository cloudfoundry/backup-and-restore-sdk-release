package mysql

import (
	"fmt"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

var _ = Describe("mysql", func() {
	var dbDumpPath string
	var configPath string
	var databaseName string
	var brJob, dbJob JobInstance
	var mysqlHostName string
	var mysqlPassword string

	BeforeEach(func() {
		mysqlHostName = MustHaveEnv("MYSQL_HOSTNAME")
		mysqlPassword = MustHaveEnv("MYSQL_PASSWORD")

		configPath = "/tmp/config.json" + strconv.FormatInt(time.Now().Unix(), 10)
		dbDumpPath = "/tmp/sql_dump" + strconv.FormatInt(time.Now().Unix(), 10)
		databaseName = "db" + strconv.FormatInt(time.Now().Unix(), 10)
	})

	Context("when the mysql server version matches", func() {
		BeforeEach(func() {
			brJob = JobInstance{
				deployment:    "mysql-dev",
				instance:      "database-backup-restorer",
				instanceIndex: "0",
			}

			dbJob = JobInstance{
				deployment:    "mysql-dev",
				instance:      "mysql",
				instanceIndex: "0",
			}

			connection, err := sql.Open("mysql", fmt.Sprintf("root:%s@tcp(%s:3306)/", mysqlPassword, mysqlHostName))
			Expect(err).NotTo(HaveOccurred())

			_, err = connection.Exec("CREATE DATABASE %s", databaseName)
			Expect(err).NotTo(HaveOccurred())

			_, err = connection.Exec("USE %s", databaseName)
			Expect(err).NotTo(HaveOccurred())

			_, err = connection.Exec("CREATE TABLE people (name varchar(255));")
			Expect(err).NotTo(HaveOccurred())

			dbJob.runMysqlSqlCommandOnDatabase(databaseName, "INSERT INTO people VALUES ('Old Person');")
			dbJob.runMysqlSqlCommandOnDatabase(databaseName, "CREATE TABLE places (name varchar(255));")
			dbJob.runMysqlSqlCommandOnDatabase(databaseName, "INSERT INTO places VALUES ('Old Place');")

			configJson := fmt.Sprintf(
				`{"username":"root","password":"%s","host":"%s","port":3306,"database":"%s","adapter":"mysql"}`,
				MustHaveEnv("MYSQL_PASSWORD"),
				mysqlHostName,
				databaseName,
			)
			brJob.runOnVMAndSucceed(fmt.Sprintf("echo '%s' > %s", configJson, configPath))
		})

		AfterEach(func() {
			dbJob.runMysqlSqlCommand(fmt.Sprintf(`echo 'DROP DATABASE %s;'`, databaseName))
			brJob.runOnVMAndSucceed(fmt.Sprintf("rm -rf %s %s", configPath, dbDumpPath))
		})

		FIt("backs up and restores the database", func() {
			brJob.runOnVMAndSucceed(fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/backup --artifact-file %s --config %s", dbDumpPath, configPath))

			dbJob.runMysqlSqlCommandOnDatabase(databaseName, "UPDATE people SET NAME = 'New Person';")
			dbJob.runMysqlSqlCommandOnDatabase(databaseName, "UPDATE places SET NAME = 'New Place';")

			brJob.runOnVMAndSucceed(fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/restore --artifact-file %s --config %s", dbDumpPath, configPath))

			Expect(dbJob.runMysqlSqlCommandOnDatabase(databaseName, "SELECT name FROM people;")).To(gbytes.Say("Old Person"))
			Expect(dbJob.runMysqlSqlCommandOnDatabase(databaseName, "SELECT name FROM people;")).NotTo(gbytes.Say("New Person"))
			Expect(dbJob.runMysqlSqlCommandOnDatabase(databaseName, "SELECT name FROM places;")).To(gbytes.Say("Old Place"))
			Expect(dbJob.runMysqlSqlCommandOnDatabase(databaseName, "SELECT name FROM places;")).NotTo(gbytes.Say("New Place"))
		})

		Context("and some existing 'tables' are specified in config", func() {
			BeforeEach(func() {
				configJson := fmt.Sprintf(
					`{"username":"root","password":"%s","host":"%s","port":3306,"database":"%s","adapter":"mysql","tables":["people"]}`,
					MustHaveEnv("MYSQL_PASSWORD"),
					dbJob.getIPOfInstance(),
					databaseName,
				)
				brJob.runOnVMAndSucceed(fmt.Sprintf("echo '%s' > %s", configJson, configPath))
			})

			It("backs up and restores only the specified tables", func() {
				brJob.runOnVMAndSucceed(fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/backup --artifact-file %s --config %s", dbDumpPath, configPath))

				dbJob.runMysqlSqlCommandOnDatabase(databaseName, "UPDATE people SET NAME = 'New Person';")
				dbJob.runMysqlSqlCommandOnDatabase(databaseName, "UPDATE places SET NAME = 'New Place';")

				brJob.runOnVMAndSucceed(fmt.Sprintf("cat %s", dbDumpPath))

				restoreSession := brJob.runOnVMAndSucceed(fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/restore --artifact-file %s --config %s", dbDumpPath, configPath))

				Expect(restoreSession).To(gbytes.Say("CREATE TABLE `people`"))

				Expect(dbJob.runMysqlSqlCommandOnDatabase(databaseName, "SELECT name FROM people;")).To(gbytes.Say("Old Person"))
				Expect(dbJob.runMysqlSqlCommandOnDatabase(databaseName, "SELECT name FROM people;")).NotTo(gbytes.Say("New Person"))
				Expect(dbJob.runMysqlSqlCommandOnDatabase(databaseName, "SELECT name FROM places;")).To(gbytes.Say("New Place"))
				Expect(dbJob.runMysqlSqlCommandOnDatabase(databaseName, "SELECT name FROM places;")).NotTo(gbytes.Say("Old Place"))
			})
		})

		Context("and 'tables' are specified in config only some of which exist", func() {
			BeforeEach(func() {
				configJson := fmt.Sprintf(
					`{"username":"root","password":"%s","host":"%s","port":3306,"database":"%s","adapter":"mysql","tables":["people", "not there"]}`,
					MustHaveEnv("MYSQL_PASSWORD"),
					dbJob.getIPOfInstance(),
					databaseName,
				)
				brJob.runOnVMAndSucceed(fmt.Sprintf("echo '%s' > %s", configJson, configPath))
			})

			It("raises an error about the non-existent tables", func() {
				session := brJob.runOnInstance(fmt.Sprintf(
					"/var/vcap/jobs/database-backup-restorer/bin/backup --artifact-file %s --config %s",
					dbDumpPath,
					configPath))

				Expect(session.ExitCode()).NotTo(BeZero())
				Expect(session).To(gbytes.Say("mysqldump: Couldn't find table: \"not there\""))
				Expect(session).To(gbytes.Say(
					"You may need to delete the artifact-file that was created before re-running"))
			})
		})

		Context("and 'tables' are specified in config none of them exist", func() {
			BeforeEach(func() {
				configJson := fmt.Sprintf(
					`{"username":"root","password":"%s","host":"%s","port":3306,"database":"%s","adapter":"mysql","tables":["lizards", "form-shifting-people"]}`,
					MustHaveEnv("MYSQL_PASSWORD"),
					dbJob.getIPOfInstance(),
					databaseName,
				)
				brJob.runOnVMAndSucceed(fmt.Sprintf("echo '%s' > %s", configJson, configPath))
			})

			It("raises an error about the non-existent tables", func() {
				session := brJob.runOnInstance(fmt.Sprintf(
					"/var/vcap/jobs/database-backup-restorer/bin/backup --artifact-file %s --config %s",
					dbDumpPath,
					configPath))

				Expect(session.ExitCode()).NotTo(BeZero())
				Expect(session).To(gbytes.Say("mysqldump: Couldn't find table: \"lizards\""))
				Expect(session).To(gbytes.Say(
					"You may need to delete the artifact-file that was created before re-running"))
			})
		})
	})
})
