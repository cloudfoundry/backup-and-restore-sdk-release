package system_tests

import (
	"io"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

var _ = Describe("backup", func() {
	Context("database-backuper is colocated with Postgres", func() {
		FIt("backs up the Postgres database", func() {
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

func RunOnInstance(deployment, instanceName, instanceIndex string, cmd ...string) *gexec.Session {
	return RunCommand(
		join(
			BoshCommand(),
			forDeployment(deployment),
			getSSHCommand(instanceName, instanceIndex),
		),
		join(cmd...),
	)
}

func RunCommand(cmd string, args ...string) *gexec.Session {
	return RunCommandWithStream(GinkgoWriter, GinkgoWriter, cmd, args...)
}

func RunCommandWithStream(stdout, stderr io.Writer, cmd string, args ...string) *gexec.Session {
	cmdParts := strings.Split(cmd, " ")
	commandPath := cmdParts[0]
	combinedArgs := append(cmdParts[1:], args...)
	command := exec.Command(commandPath, combinedArgs...)

	session, err := gexec.Start(command, stdout, stderr)

	Expect(err).ToNot(HaveOccurred())
	Eventually(session).Should(gexec.Exit())
	return session
}

func BoshCommand() string {
	return fmt.Sprintf("bosh-cli --non-interactive --environment=%s --ca-cert=%s --client=%s --client-secret=%s",
		MustHaveEnv("BOSH_URL"),
		MustHaveEnv("BOSH_CERT_PATH"),
		MustHaveEnv("BOSH_CLIENT"),
		MustHaveEnv("BOSH_CLIENT_SECRET"),
	)
}

func forDeployment(deploymentName string) string {
	return fmt.Sprintf(
		"--deployment=%s",
		deploymentName,
	)
}

func getSSHCommand(instanceName, instanceIndex string) string {
	return fmt.Sprintf(
		"ssh --gw-user=%s --gw-host=%s --gw-private-key=%s %s/%s",
		MustHaveEnv("BOSH_GATEWAY_USER"),
		MustHaveEnv("BOSH_GATEWAY_HOST"),
		MustHaveEnv("BOSH_GATEWAY_KEY"),
		instanceName,
		instanceIndex,
	)
}

func MustHaveEnv(keyname string) string {
	val := os.Getenv(keyname)
	Expect(val).NotTo(BeEmpty(), "Need "+keyname+" for the test")
	return val
}

func join(args ...string) string {
	return strings.Join(args, " ")
}
