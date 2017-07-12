package system_tests

import (
	"fmt"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("mysql-backup", func() {
	Context("database-backuper lives on its own instance", func() {
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

			configJson := fmt.Sprintf(
				`{"username":"root","password":"%s","host":"localhost","port":3306,"database":"%s","adapter":"mysql"}`,
				MustHaveEnv("MYSQL_PASSWORD"),
				databaseName,
			)

			RunOnInstance("mysql-dev", "database-backuper", "0",
				fmt.Sprintf("echo '%s' > %s", configJson, configPath))
		})

		AfterEach(func() {
			runOnMysqlVMAndSucceed(fmt.Sprintf(`echo 'DROP DATABASE %s;' | /var/vcap/packages/mariadb/bin/mysql -u root -h localhost --password='%s'`, databaseName, MustHaveEnv("MYSQL_PASSWORD")))
			RunOnInstance("mysql-dev", "database-backuper", "0",
				fmt.Sprintf("rm -rf %s %s", configPath, dbDumpPath))
		})

		It("backs up the MySQL database", func() {
			backupSession := RunOnInstance("mysql-dev", "database-backuper", "0",
				fmt.Sprintf("/var/vcap/jobs/database-backuper/bin/backup --artifact-file %s --config %s", dbDumpPath, configPath))
			Expect(backupSession).To(gexec.Exit(0))

			fileCheckSession := RunOnInstance("mysql-dev", "database-backuper", "0",
				fmt.Sprintf("ls -l %s", dbDumpPath))
			Expect(fileCheckSession).To(gexec.Exit(0))
		})
	})

})
