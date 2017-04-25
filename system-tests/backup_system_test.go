package system_tests

import (
	"io"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"fmt"
	"os"
	"os/exec"
	"strings"
)

var _ = Describe("backup", func() {
	It("backs up a Postgres database", func() {
		expectFilename := "/tmp/sql_dump"
		Expect(RunOnInstance("postgres-dev", "postgres", "0",
			fmt.Sprintf("rm -rf %s", expectFilename))).To(gexec.Exit(0))
		Expect(RunOnInstance("postgres-dev", "postgres", "0",
			fmt.Sprintf("HOST=localhost PORT=5432 USER=bosh PASSWORD=%s DATABASE=bosh OUTPUT_FILE=%s /var/vcap/jobs/database_backuper/bin/backup",
				MustHaveEnv("POSTGRES_PASSWORD"), expectFilename))).To(gexec.Exit(0))
		Expect(RunOnInstance("postgres-dev", "postgres", "0",
			fmt.Sprintf("ls -l %s", expectFilename))).To(gexec.Exit(0))
	})
})

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
