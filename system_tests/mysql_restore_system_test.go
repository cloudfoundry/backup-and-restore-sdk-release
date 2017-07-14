package system_tests

import (
	"fmt"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("mysql-restore", func() {
	var dbDumpPath string
	var configPath string
	var databaseName string

	BeforeEach(func() {
		configPath = "/tmp/config.json" + strconv.FormatInt(time.Now().Unix(), 10)
		dbDumpPath = "/tmp/sql_dump" + strconv.FormatInt(time.Now().Unix(), 10)
		databaseName = "db" + strconv.FormatInt(time.Now().Unix(), 10)

		runOnMysqlVMAndSucceed(fmt.Sprintf(`echo 'CREATE DATABASE %s;' | /var/vcap/packages/mariadb/bin/mysql -u root -h localhost --password='%s'`, databaseName, MustHaveEnv("MYSQL_PASSWORD")))
		runMysqlSqlCommand("CREATE TABLE people (name varchar(255));", databaseName)
		runMysqlSqlCommand("INSERT INTO people VALUES ('Derik');", databaseName)

		ip := getIPOfInstance("mysql-dev", "mysql")
		configJson := fmt.Sprintf(
			`{"username":"root","password":"%s","host":"%s","port":3306,"database":"%s","adapter":"mysql"}`,
			MustHaveEnv("MYSQL_PASSWORD"),
			ip,
			databaseName,
		)

		RunOnInstance("mysql-dev", "database-backup-restorer", "0",
			fmt.Sprintf("echo '%s' > %s", configJson, configPath))
	})

	AfterEach(func() {
		runOnMysqlVMAndSucceed(fmt.Sprintf(`echo 'DROP DATABASE %s;' | /var/vcap/packages/mariadb/bin/mysql -u root -h localhost --password='%s'`, databaseName, MustHaveEnv("MYSQL_PASSWORD")))
		RunOnInstance("mysql-dev", "database-backup-restorer", "0",
			fmt.Sprintf("rm -rf %s %s", configPath, dbDumpPath))
	})

	Context("database-backup-restorer lives on its own instance", func() {
		It("restores the MySQL database", func() {
			backupSession := RunOnInstance("mysql-dev", "database-backup-restorer", "0",
				fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/backup --artifact-file %s --config %s", dbDumpPath, configPath))
			Expect(backupSession).To(gexec.Exit(0))

			runMysqlSqlCommand("UPDATE people SET NAME = 'Dave';", databaseName)

			restoreSession := RunOnInstance("mysql-dev", "database-backup-restorer", "0",
				fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/restore --artifact-file %s --config %s", dbDumpPath, configPath))
			Expect(restoreSession).To(gexec.Exit(0))

			Expect(runMysqlSqlCommand("SELECT name FROM people;", databaseName)).To(gbytes.Say("Derik"))
			Expect(runMysqlSqlCommand("SELECT name FROM people;", databaseName)).NotTo(gbytes.Say("Dave"))
		})
	})
})
