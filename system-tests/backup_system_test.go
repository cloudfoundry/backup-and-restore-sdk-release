package system_tests

import (
	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"encoding/json"
	"fmt"
	"strings"
	"time"
)

var _ = Describe("backup", func() {
	Context("database-backuper is colocated with Postgres", func() {
		var dbDumpPath string
		var configPath string
		var databaseName string

		BeforeEach(func() {
			configPath = "/tmp/config.json" + strconv.FormatInt(time.Now().Unix(), 10)
			dbDumpPath = "/tmp/sql_dump" + strconv.FormatInt(time.Now().Unix(), 10)
			databaseName = "db" + strconv.FormatInt(time.Now().Unix(), 10)

			runOnPostgresVMAndSucceed(fmt.Sprintf(`/var/vcap/packages/postgres-9.4/bin/createdb -U vcap "%s"`, databaseName))
			runSqlCommand("CREATE TABLE people (name varchar);", databaseName)
			runSqlCommand("INSERT INTO people VALUES ('Derik');", databaseName)

			configJson := fmt.Sprintf(
				`{"username":"vcap","password":"%s","host":"localhost","port":"5432","database":"%s","adapter":"postgres"}`,
				MustHaveEnv("POSTGRES_PASSWORD"),
				databaseName,
			)

			runOnPostgresVMAndSucceed(fmt.Sprintf("echo '%s' > %s", configJson, configPath))
		})

		AfterEach(func() {
			runOnPostgresVMAndSucceed(fmt.Sprintf(`/var/vcap/packages/postgres-9.4/bin/dropdb -U vcap "%s"`, databaseName))
			runOnPostgresVMAndSucceed(fmt.Sprintf("rm -rf %s %s", configPath, dbDumpPath))
		})

		It("backs up the Postgres database", func() {
			runOnPostgresVMAndSucceed(fmt.Sprintf("/var/vcap/jobs/database-backuper/bin/backup --artifact-file %s --config %s", dbDumpPath, configPath))
			runOnPostgresVMAndSucceed(fmt.Sprintf("ls -l %s", dbDumpPath))
		})
	})

	Context("database-backuper lives on its own instance", func() {
		It("backs up the Postgres database", func() {
			expectFilename := "/tmp/sql_dump"

			deploymentName := "postgres-dev"
			ip := getIPOfInstance(deploymentName, "postgres")
			configJson := fmt.Sprintf(
				`{"username":"bosh","password":"%s","host":"%s","port":"5432","database":"bosh","adapter":"postgres"}`,
				MustHaveEnv("POSTGRES_PASSWORD"),
				ip,
			)

			Expect(RunOnInstance(deploymentName, "database-backuper", "0",
				fmt.Sprintf("rm -rf /tmp/config.json %s", expectFilename))).To(gexec.Exit(0))
			Expect(RunOnInstance(deploymentName, "database-backuper", "0",
				fmt.Sprintf("echo '%s' >> /tmp/config.json", configJson))).To(gexec.Exit(0))

			Expect(RunOnInstance(deploymentName, "database-backuper", "0",
				fmt.Sprintf("/var/vcap/jobs/database-backuper/bin/backup --artifact-file %s --config /tmp/config.json", expectFilename),
			)).To(gexec.Exit(0))
			Expect(RunOnInstance(deploymentName, "database-backuper", "0",
				fmt.Sprintf("ls -l %s", expectFilename))).To(gexec.Exit(0))
		})
	})
})

func getIPOfInstance(deploymentName, instanceName string) string {
	session := RunCommand(
		BoshCommand(),
		forDeployment(deploymentName),
		"instances",
		"--json",
	)
	outputFromCli := jsonOutputFromCli{}
	contents := session.Out.Contents()
	Expect(json.Unmarshal(contents, &outputFromCli)).To(Succeed())
	for _, instanceData := range outputFromCli.Tables[0].Rows {
		if strings.HasPrefix(instanceData["instance"], instanceName+"/") {
			return instanceData["ips"]
		}
	}
	Fail("Cant find instances with name '" + instanceName + "' and deployment name '" + deploymentName + "'")
	return ""
}

type jsonOutputFromCli struct {
	Tables []struct {
		Rows []map[string]string
	}
}
