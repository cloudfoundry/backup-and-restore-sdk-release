package mysql

import (
	"fmt"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"database/sql"

	"os/exec"

	_ "github.com/go-sql-driver/mysql"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("mysql", func() {
	var dbDumpPath string
	var configPath string
	var databaseName string
	var brJob JobInstance
	var mysqlHostName string
	var proxySession *gexec.Session
	var connection *sql.DB

	BeforeEach(func() {
		configPath = "/tmp/config.json" + strconv.FormatInt(time.Now().Unix(), 10)
		dbDumpPath = "/tmp/sql_dump" + strconv.FormatInt(time.Now().Unix(), 10)
		databaseName = "db" + strconv.FormatInt(time.Now().Unix(), 10)
	})
	BeforeSuite(func() {
		mysqlHostName = MustHaveEnv("MYSQL_HOSTNAME")
		connection, proxySession = connect()
	})
	AfterSuite(func() {
		if proxySession != nil {
			proxySession.Kill()
		}
	})

	Context("when the mysql server version matches", func() {
		BeforeEach(func() {
			brJob = JobInstance{
				deployment:    "mysql-dev",
				instance:      "database-backup-restorer",
				instanceIndex: "0",
			}

			runSQLCommand("CREATE DATABASE "+databaseName, connection)

			runSQLCommand("USE "+databaseName, connection)

			runSQLCommand("CREATE TABLE people (name varchar(255));", connection)

			runSQLCommand("INSERT INTO people VALUES ('Old Person');", connection)
			runSQLCommand("CREATE TABLE places (name varchar(255));", connection)
			runSQLCommand("INSERT INTO places VALUES ('Old Place');", connection)

			configJson := fmt.Sprintf(
				`{"username":"%s","password":"%s","host":"%s","port":3306,"database":"%s","adapter":"mysql"}`,
				MustHaveEnv("MYSQL_USERNAME"),
				MustHaveEnv("MYSQL_PASSWORD"),
				mysqlHostName,
				databaseName,
			)
			brJob.runOnVMAndSucceed(fmt.Sprintf("echo '%s' > %s", configJson, configPath))
		})

		AfterEach(func() {
			runSQLCommand("DROP DATABASE "+databaseName, connection)
			brJob.runOnVMAndSucceed(fmt.Sprintf("rm -rf %s %s", configPath, dbDumpPath))
		})

		It("backs up and restores the database", func() {
			brJob.runOnVMAndSucceed(fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/backup --artifact-file %s --config %s", dbDumpPath, configPath))

			runSQLCommand("UPDATE people SET NAME = 'New Person';", connection)
			runSQLCommand("UPDATE places SET NAME = 'New Place';", connection)

			brJob.runOnVMAndSucceed(fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/restore --artifact-file %s --config %s", dbDumpPath, configPath))

			Expect(fetchSQLColumn("SELECT name FROM people;", connection)).To(
				ConsistOf("Old Person"))
			Expect(fetchSQLColumn("SELECT name FROM people;", connection)).NotTo(
				ConsistOf("New Person"))
			Expect(fetchSQLColumn("SELECT name FROM places;", connection)).To(
				ConsistOf("Old Place"))
			Expect(fetchSQLColumn("SELECT name FROM places;", connection)).NotTo(
				ConsistOf("New Place"))
		})

		Context("and some existing 'tables' are specified in config", func() {
			BeforeEach(func() {
				configJson := fmt.Sprintf(
					`{"username":"%s","password":"%s","host":"%s","port":3306,"database":"%s","adapter":"mysql","tables":["people"]}`,
					MustHaveEnv("MYSQL_USERNAME"),
					MustHaveEnv("MYSQL_PASSWORD"),
					mysqlHostName,
					databaseName,
				)
				brJob.runOnVMAndSucceed(fmt.Sprintf("echo '%s' > %s", configJson, configPath))
			})

			It("backs up and restores only the specified tables", func() {
				brJob.runOnVMAndSucceed(fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/backup --artifact-file %s --config %s", dbDumpPath, configPath))

				runSQLCommand("UPDATE people SET NAME = 'New Person';", connection)
				runSQLCommand("UPDATE places SET NAME = 'New Place';", connection)

				brJob.runOnVMAndSucceed(fmt.Sprintf("cat %s", dbDumpPath))

				restoreSession := brJob.runOnVMAndSucceed(fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/restore --artifact-file %s --config %s", dbDumpPath, configPath))

				Expect(restoreSession).To(gbytes.Say("CREATE TABLE `people`"))

				Expect(fetchSQLColumn("SELECT name FROM people;", connection)).To(
					ConsistOf("Old Person"))
				Expect(fetchSQLColumn("SELECT name FROM people;", connection)).NotTo(
					ConsistOf("New Person"))
				Expect(fetchSQLColumn("SELECT name FROM places;", connection)).To(
					ConsistOf("New Place"))
				Expect(fetchSQLColumn("SELECT name FROM places;", connection)).NotTo(
					ConsistOf("Old Place"))
			})
		})

		Context("and 'tables' are specified in config only some of which exist", func() {
			BeforeEach(func() {
				configJson := fmt.Sprintf(
					`{"username":"%s","password":"%s","host":"%s","port":3306,"database":"%s","adapter":"mysql","tables":["people", "not there"]}`,
					MustHaveEnv("MYSQL_USERNAME"),
					MustHaveEnv("MYSQL_PASSWORD"),
					mysqlHostName,
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
					`{"username":"%s","password":"%s","host":"%s","port":3306,"database":"%s","adapter":"mysql","tables":["lizards", "form-shifting-people"]}`,
					MustHaveEnv("MYSQL_USERNAME"),
					MustHaveEnv("MYSQL_PASSWORD"),
					mysqlHostName,
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

func connect() (*sql.DB, *gexec.Session) {
	mysqlHostName := MustHaveEnv("MYSQL_HOSTNAME")
	mysqlPassword := MustHaveEnv("MYSQL_PASSWORD")
	mysqlUsername := MustHaveEnv("MYSQL_USERNAME")
	mysqlPort := MustHaveEnv("MYSQL_PORT")

	sshProxyHost := MustHaveEnv("SSH_PROXY_HOST")
	sshProxyUser := MustHaveEnv("SSH_PROXY_USER")
	sshProxyKeyFile := MustHaveEnv("SSH_PROXY_KEY_FILE")

	proxiedMysqlHostName := "127.0.0.1"
	proxiedMysqlPort := "13306"
	var err error
	proxySession, err := gexec.Start(exec.Command(
		"ssh",
		"-L",
		fmt.Sprintf("%s:%s:%s", proxiedMysqlPort, mysqlHostName, mysqlPort),
		sshProxyUser+"@"+sshProxyHost,
		"-i", sshProxyKeyFile,
		"-N",
		"-o",
		"'UserKnownHostsFile=/dev/null'",
		"-o",
		"'StrictHostKeyChecking=no'",
	), GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	time.Sleep(1 * time.Second)
	connection, err := sql.Open("mysql", fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/", mysqlUsername, mysqlPassword, proxiedMysqlHostName, proxiedMysqlPort))
	Expect(err).NotTo(HaveOccurred())

	return connection, proxySession
}

func runSQLCommand(command string, connection *sql.DB) {
	_, err := connection.Exec(command)
	Expect(err).NotTo(HaveOccurred())
}

func fetchSQLColumn(command string, connection *sql.DB) []string {
	var returnValue []string
	rows, err := connection.Query(command)
	Expect(err).NotTo(HaveOccurred())

	defer rows.Close()
	for rows.Next() {
		var rowData string
		Expect(rows.Scan(&rowData)).NotTo(HaveOccurred())

		returnValue = append(returnValue, rowData)
	}
	Expect(rows.Err()).NotTo(HaveOccurred())
	return returnValue
}
