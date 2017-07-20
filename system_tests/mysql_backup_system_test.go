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

var _ = Describe("mysql-backup", func() {
	var dbDumpPath string
	var configPath string
	var databaseName string
	var dbJob, brJob JobInstance

	BeforeEach(func() {
		configPath = "/tmp/config.json" + strconv.FormatInt(time.Now().Unix(), 10)
		dbDumpPath = "/tmp/sql_dump" + strconv.FormatInt(time.Now().Unix(), 10)
		databaseName = "db" + strconv.FormatInt(time.Now().Unix(), 10)
	})

	JustBeforeEach(func() {
		dbJob.runOnVMAndSucceed(fmt.Sprintf(`echo 'CREATE DATABASE %s;' | /var/vcap/packages/mariadb/bin/mysql -u root -h localhost --password='%s'`, databaseName, MustHaveEnv("MYSQL_PASSWORD")))
		dbJob.runMysqlSqlCommand("CREATE TABLE people (name varchar(255));", databaseName)
		dbJob.runMysqlSqlCommand("INSERT INTO people VALUES ('Derik');", databaseName)

		ip := dbJob.getIPOfInstance()

		configJson := fmt.Sprintf(
			`{"username":"root","password":"%s","host":"%s","port":3306,"database":"%s","adapter":"mysql"}`,
			MustHaveEnv("MYSQL_PASSWORD"),
			ip,
			databaseName,
		)

		brJob.RunOnInstance(fmt.Sprintf("echo '%s' > %s", configJson, configPath))
	})

	AfterEach(func() {
		dbJob.runOnVMAndSucceed(fmt.Sprintf(`echo 'DROP DATABASE %s;' | /var/vcap/packages/mariadb/bin/mysql -u root -h localhost --password='%s'`, databaseName, MustHaveEnv("MYSQL_PASSWORD")))
		brJob.RunOnInstance(fmt.Sprintf("rm -rf %s %s", configPath, dbDumpPath))
	})

	Context("database-backup-restorer lives on its own instance", func() {
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
		})

		It("backs up the MySQL database", func() {
			backupSession := brJob.RunOnInstance(fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/backup --artifact-file %s --config %s", dbDumpPath, configPath))
			Expect(backupSession).To(gexec.Exit(0))

			fileCheckSession := brJob.RunOnInstance(fmt.Sprintf("ls -l %s", dbDumpPath))
			Expect(fileCheckSession).To(gexec.Exit(0))
		})
	})

	Context("mysql server is different version", func() {
		BeforeEach(func() {
			brJob = JobInstance{
				deployment:    "mysql-old-dev",
				instance:      "database-backup-restorer",
				instanceIndex: "0",
			}

			dbJob = JobInstance{
				deployment:    "mysql-old-dev",
				instance:      "mysql",
				instanceIndex: "0",
			}
		})

		It("fails with a helpful message", func() {
			backupSession := brJob.RunOnInstance(fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/backup --artifact-file %s --config %s", dbDumpPath, configPath))
			Expect(backupSession).To(gexec.Exit(1))
			Expect(backupSession).To(gbytes.Say("major/minor version mismatch"))

		})
	})
})
