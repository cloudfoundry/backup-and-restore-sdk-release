package system_tests

import (
	"encoding/json"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"strings"
)

type JobInstance struct {
	deployment    string
	instance      string
	instanceIndex string
}

func (jobInstance *JobInstance) runPostgresSqlCommand(command, database string) *gexec.Session {
	return jobInstance.runOnVMAndSucceed(
		fmt.Sprintf(`/var/vcap/packages/postgres-9.4/bin/psql -U vcap "%s" --command="%s"`, database, command),
	)
}

func (jobInstance *JobInstance) runMysqlSqlCommand(command, database string) *gexec.Session {
	return jobInstance.runOnVMAndSucceed(
		fmt.Sprintf(`echo -e "%s" | /var/vcap/packages/mariadb/bin/mysql -u root -h localhost --password='%s' "%s"`, command, MustHaveEnv("MYSQL_PASSWORD"), database),
	)
}

func (jobInstance *JobInstance) runOnVMAndSucceed(command string) *gexec.Session {
	session := jobInstance.RunOnInstance(command)
	Expect(session).To(gexec.Exit(0))

	return session
}

func (jobInstance *JobInstance) RunOnInstance(cmd ...string) *gexec.Session {
	return RunCommand(
		join(
			BoshCommand(),
			forDeployment(jobInstance.deployment),
			getSSHCommand(jobInstance.instance, jobInstance.instanceIndex),
		),
		join(cmd...),
	)
}

func (jobInstance *JobInstance) getIPOfInstance() string {
	session := RunCommand(
		BoshCommand(),
		forDeployment(jobInstance.deployment),
		"instances",
		"--json",
	)
	outputFromCli := jsonOutputFromCli{}
	contents := session.Out.Contents()
	Expect(json.Unmarshal(contents, &outputFromCli)).To(Succeed())
	for _, instanceData := range outputFromCli.Tables[0].Rows {
		if strings.HasPrefix(instanceData["instance"], jobInstance.instance+"/") {
			return instanceData["ips"]
		}
	}
	Fail("Cant find instances with name '" + jobInstance.instance + "' and deployment name '" + jobInstance.deployment + "'")
	return ""
}
