package mysql

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	_ "github.com/go-sql-driver/mysql"

	. "database-backup-restore/system_tests/utils"
)

var _ = Describe("mysql", func() {
	var dbDumpPath string
	var configPath string
	var databaseName string

	BeforeEach(func() {
		disambiguationString := DisambiguationString()
		configPath = "/tmp/config" + disambiguationString
		dbDumpPath = "/tmp/artifact" + disambiguationString
		databaseName = "db" + disambiguationString
	})

	Context("when the mysql server version matches", func() {
		BeforeEach(func() {
			RunSQLCommand("CREATE DATABASE "+databaseName, connection)

			RunSQLCommand("USE "+databaseName, connection)

			RunSQLCommand("CREATE TABLE people (name varchar(255));", connection)

			RunSQLCommand("INSERT INTO people VALUES ('Old Person');", connection)
			RunSQLCommand("CREATE TABLE places (name varchar(255));", connection)
			RunSQLCommand("INSERT INTO places VALUES ('Old Place');", connection)
		})

		AfterEach(func() {
			RunSQLCommand("DROP DATABASE "+databaseName, connection)
			brJob.RunOnVMAndSucceed(fmt.Sprintf("sudo rm -rf %s %s", configPath, dbDumpPath))
		})

		Context("when we backup the whole database", func() {
			BeforeEach(func() {
				configJson := fmt.Sprintf(
					`{"username":"%s","password":"%s","host":"%s","port":%d,"database":"%s","adapter":"mysql"}`,
					mysqlNonSslUsername,
					mysqlPassword,
					mysqlHostName,
					mysqlPort,
					databaseName,
				)
				brJob.RunOnVMAndSucceed(fmt.Sprintf("echo '%s' > %s", configJson, configPath))
				brJob.RunOnVMAndSucceed(
					fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/backup --artifact-file %s --config %s",
						dbDumpPath, configPath))

			})

			It("restores the database successfully", func() {
				RunSQLCommand("UPDATE people SET NAME = 'New Person';", connection)
				RunSQLCommand("UPDATE places SET NAME = 'New Place';", connection)

				brJob.RunOnVMAndSucceed(
					fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/restore --artifact-file %s --config %s",
						dbDumpPath, configPath))

				Expect(FetchSQLColumn("SELECT name FROM people;", connection)).To(
					ConsistOf("Old Person"))
				Expect(FetchSQLColumn("SELECT name FROM people;", connection)).NotTo(
					ConsistOf("New Person"))
				Expect(FetchSQLColumn("SELECT name FROM places;", connection)).To(
					ConsistOf("Old Place"))
				Expect(FetchSQLColumn("SELECT name FROM places;", connection)).NotTo(
					ConsistOf("New Place"))
			})

			Context("when tables do not exist", func() {
				It("restores the tables successfully", func() {
					RunSQLCommand("DROP TABLE people;", connection)

					brJob.RunOnVMAndSucceed(
						fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/restore --config %s --artifact-file %s",
							configPath, dbDumpPath))

					Expect(FetchSQLColumn("SELECT name FROM people;", connection)).
						To(ConsistOf("Old Person"))
				})
			})

		})

		Context("when some existing 'tables' are specified in config", func() {
			BeforeEach(func() {
				configJson := fmt.Sprintf(
					`{"username":"%s","password":"%s","host":"%s","port":%d,"database":"%s","adapter":"mysql","tables":["people"]}`,
					mysqlNonSslUsername,
					mysqlPassword,
					mysqlHostName,
					mysqlPort,
					databaseName,
				)
				brJob.RunOnVMAndSucceed(fmt.Sprintf("echo '%s' > %s", configJson, configPath))
			})

			It("backs up and restores only the specified tables", func() {
				brJob.RunOnVMAndSucceed(
					fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/backup --artifact-file %s --config %s",
						dbDumpPath, configPath))

				RunSQLCommand("UPDATE people SET NAME = 'New Person';", connection)
				RunSQLCommand("UPDATE places SET NAME = 'New Place';", connection)

				brJob.RunOnVMAndSucceed(fmt.Sprintf("cat %s", dbDumpPath))

				restoreSession := brJob.RunOnVMAndSucceed(
					fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/restore --artifact-file %s --config %s",
						dbDumpPath, configPath))

				Expect(restoreSession).To(gbytes.Say("CREATE TABLE `people`"))

				Expect(FetchSQLColumn("SELECT name FROM people;", connection)).To(
					ConsistOf("Old Person"))
				Expect(FetchSQLColumn("SELECT name FROM people;", connection)).NotTo(
					ConsistOf("New Person"))
				Expect(FetchSQLColumn("SELECT name FROM places;", connection)).To(
					ConsistOf("New Place"))
				Expect(FetchSQLColumn("SELECT name FROM places;", connection)).NotTo(
					ConsistOf("Old Place"))
			})
		})

		Context("when 'tables' are specified in config only some of which exist", func() {
			BeforeEach(func() {
				configJson := fmt.Sprintf(
					`{"username":"%s","password":"%s","host":"%s","port":%d,"database":"%s","adapter":"mysql","tables":["people", "not there"]}`,
					mysqlNonSslUsername,
					mysqlPassword,
					mysqlHostName,
					mysqlPort,
					databaseName,
				)
				brJob.RunOnVMAndSucceed(fmt.Sprintf("echo '%s' > %s", configJson, configPath))
			})

			It("raises an error about the non-existent tables", func() {
				session := brJob.RunOnInstance(fmt.Sprintf(
					"/var/vcap/jobs/database-backup-restorer/bin/backup --artifact-file %s --config %s",
					dbDumpPath,
					configPath))

				Expect(session.ExitCode()).NotTo(BeZero())
				Expect(session).To(gbytes.Say("mysqldump: Couldn't find table: \"not there\""))
				Expect(session).To(gbytes.Say(
					"You may need to delete the artifact-file that was created before re-running"))
			})
		})

		Context("when 'tables' are specified in config none of them exist", func() {
			BeforeEach(func() {
				configJson := fmt.Sprintf(
					`{"username":"%s","password":"%s","host":"%s","port":%d,"database":"%s","adapter":"mysql","tables":["lizards", "form-shifting-people"]}`,
					mysqlNonSslUsername,
					mysqlPassword,
					mysqlHostName,
					mysqlPort,
					databaseName,
				)
				brJob.RunOnVMAndSucceed(fmt.Sprintf("echo '%s' > %s", configJson, configPath))
			})

			It("raises an error about the non-existent tables", func() {
				session := brJob.RunOnInstance(fmt.Sprintf(
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
