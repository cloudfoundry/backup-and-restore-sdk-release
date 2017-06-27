package system_tests

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"encoding/json"
	"fmt"
	"strings"
)

var _ = Describe("backup", func() {
	Context("database-backuper is colocated with Postgres", func() {
		It("backs up the Postgres database", func() {
			expectFilename := "/tmp/sql_dump"
			configJson := fmt.Sprintf(
				`{"username":"bosh","password":"%s","host":"localhost","port":"5432","database":"bosh","adapter":"postgres","output_file":"%s"}`,
				MustHaveEnv("POSTGRES_PASSWORD"),
				expectFilename,
			)
			Expect(RunOnInstance("postgres-dev", "postgres", "0",
				fmt.Sprintf("rm -rf /tmp/config.json %s", expectFilename))).To(gexec.Exit(0))
			Expect(RunOnInstance("postgres-dev", "postgres", "0",
				fmt.Sprintf("echo '%s' >> /tmp/config.json", configJson))).To(gexec.Exit(0))
			Expect(
				RunOnInstance("postgres-dev", "postgres", "0",
					fmt.Sprintf("/var/vcap/jobs/database-backuper/bin/backup /tmp/config.json"),
				)).To(gexec.Exit(0))
			Expect(RunOnInstance("postgres-dev", "postgres", "0",
				fmt.Sprintf("ls -l %s", expectFilename))).To(gexec.Exit(0))
		})
	})

	Context("database-backuper lives on its own instance", func() {
		It("backs up the Postgres database", func() {
			expectFilename := "/tmp/sql_dump"

			ip := getIPOfInstance("postgres-dev", "postgres")

			configJson := fmt.Sprintf(
				`{"username":"bosh","password":"%s","host":"%s","port":"5432","database":"bosh","adapter":"postgres","output_file":"%s"}`,
				MustHaveEnv("POSTGRES_PASSWORD"),
				ip,
				expectFilename,
			)

			Expect(RunOnInstance("postgres-dev", "database-backuper", "0",
				fmt.Sprintf("rm -rf /tmp/config.json %s", expectFilename))).To(gexec.Exit(0))
			Expect(RunOnInstance("postgres-dev", "database-backuper", "0",
				fmt.Sprintf("echo '%s' >> /tmp/config.json", configJson))).To(gexec.Exit(0))

			Expect(RunOnInstance("postgres-dev", "database-backuper", "0",
				fmt.Sprintf("/var/vcap/jobs/database-backuper/bin/backup /tmp/config.json"),
			)).To(gexec.Exit(0))
			Expect(RunOnInstance("postgres-dev", "database-backuper", "0",
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
	thing := jsonOutputFromCli{}
	contents := session.Out.Contents()
	Expect(json.Unmarshal(contents, &thing)).To(Succeed())
	for _, instanceData := range thing.Tables[0].Rows {
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
