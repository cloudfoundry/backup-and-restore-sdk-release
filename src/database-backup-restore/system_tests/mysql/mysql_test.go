package mysql

import (
	"fmt"
	"os/exec"

	_ "github.com/go-sql-driver/mysql"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "database-backup-restore/system_tests/utils"
)

var _ = Describe("mysql", func() {
	var (
		dbDumpPath   string
		configPath   string
		databaseName string
	)

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
			// brJob.RunOnVMAndSucceed(fmt.Sprintf("sudo rm -rf %s %s", configPath, dbDumpPath))
			exec.Command(fmt.Sprintf("sudo rm -rf %s %s", configPath, dbDumpPath)).CombinedOutput()
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
				// brJob.RunOnVMAndSucceed(fmt.Sprintf("echo '%s' > %s", configJson, configPath))
				exec.Command("bash", "-c", fmt.Sprintf("echo '%s' > %s", configJson, configPath)).CombinedOutput()
				// brJob.RunOnVMAndSucceed(
				//	fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/backup --artifact-file %s --config %s",
				//		dbDumpPath, configPath))
				exec.Command("bash", "-c",
					fmt.Sprintf("database-backup-restore --backup --artifact-file %s --config %s",
						dbDumpPath, configPath)).CombinedOutput()

			})

			It("restores the database successfully", func() {
				RunSQLCommand("UPDATE people SET NAME = 'New Person';", connection)
				RunSQLCommand("UPDATE places SET NAME = 'New Place';", connection)

				// brJob.RunOnVMAndSucceed(
				//	fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/restore --artifact-file %s --config %s",
				//		dbDumpPath, configPath))
				exec.Command("bash", "-c",
					fmt.Sprintf("database-backup-restore --restore --artifact-file %s --config %s",
						dbDumpPath, configPath)).CombinedOutput()

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

					// brJob.RunOnVMAndSucceed(
					//	fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/restore --config %s --artifact-file %s",
					//		configPath, dbDumpPath))
					exec.Command("bash", "-c",
						fmt.Sprintf("database-backup-restore --restore --config %s --artifact-file %s",
							configPath, dbDumpPath)).CombinedOutput()

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
				// brJob.RunOnVMAndSucceed(fmt.Sprintf("echo '%s' > %s", configJson, configPath))
				exec.Command("bash", "-c", fmt.Sprintf("echo '%s' > %s", configJson, configPath)).CombinedOutput()
			})

			It("backs up and restores only the specified tables", func() {
				//brJob.RunOnVMAndSucceed(
				//	fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/backup --artifact-file %s --config %s",
				//		dbDumpPath, configPath))
				exec.Command("bash", "-c",
					fmt.Sprintf("database-backup-restore --backup --artifact-file %s --config %s",
						dbDumpPath, configPath)).CombinedOutput()

				RunSQLCommand("UPDATE people SET NAME = 'New Person';", connection)
				RunSQLCommand("UPDATE places SET NAME = 'New Place';", connection)

				// brJob.RunOnVMAndSucceed(fmt.Sprintf("cat %s", dbDumpPath))
				exec.Command(fmt.Sprintf("cat %s", dbDumpPath)).CombinedOutput()

				// restoreSession := brJob.RunOnVMAndSucceed(
				//	fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/restore --artifact-file %s --config %s",
				//		dbDumpPath, configPath))
				restoreSession, _ := exec.Command("bash", "-c",
					fmt.Sprintf("database-backup-restore --restore --artifact-file %s --config %s",
						dbDumpPath, configPath)).CombinedOutput()

				Expect(string(restoreSession)).To(ContainSubstring("CREATE TABLE `people`"))

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
				//brJob.RunOnVMAndSucceed(fmt.Sprintf("echo '%s' > %s", configJson, configPath))
				exec.Command("bash", "-c", fmt.Sprintf("echo '%s' > %s", configJson, configPath)).CombinedOutput()
			})

			It("raises an error about the non-existent tables", func() {
				//session := brJob.RunOnInstance(fmt.Sprintf(
				//	"/var/vcap/jobs/database-backup-restorer/bin/backup --artifact-file %s --config %s",
				//	dbDumpPath,
				//	configPath))
				msg, err := exec.Command("bash", "-c", fmt.Sprintf(
					"database-backup-restore --backup --artifact-file %s --config %s",
					dbDumpPath,
					configPath)).CombinedOutput()

				Expect(err).To(HaveOccurred())
				Expect(string(msg)).To(ContainSubstring("Couldn't find table: \"not there\""))
				Expect(string(msg)).To(ContainSubstring(
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
				// brJob.RunOnVMAndSucceed(fmt.Sprintf("echo '%s' > %s", configJson, configPath))
				exec.Command("bash", "-c", fmt.Sprintf("echo '%s' > %s", configJson, configPath)).CombinedOutput()
			})

			It("raises an error about the non-existent tables", func() {
				//session := brJob.RunOnInstance(fmt.Sprintf(
				//	"/var/vcap/jobs/database-backup-restorer/bin/backup --artifact-file %s --config %s",
				//	dbDumpPath,
				//	configPath))
				msg, err := exec.Command("bash", "-c", fmt.Sprintf(
					"database-backup-restore --backup --artifact-file %s --config %s",
					dbDumpPath,
					configPath)).CombinedOutput()

				Expect(err).To(HaveOccurred())
				Expect(string(msg)).To(ContainSubstring("Couldn't find table: \"lizards\""))
				Expect(string(msg)).To(ContainSubstring(
					"You may need to delete the artifact-file that was created before re-running"))
			})
		})
	})
})
