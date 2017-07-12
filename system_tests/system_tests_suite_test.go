package system_tests

import (
	"fmt"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"encoding/json"
	"io"
	"os/exec"
	"strings"
	"testing"
)

func TestSystemTests(t *testing.T) {
	RegisterFailHandler(Fail)
	SetDefaultEventuallyTimeout(15 * time.Minute)
	RunSpecs(t, "SystemTests Suite")
}

func runPostgresSqlCommand(command, database string) *gexec.Session {
	return runOnPostgresVMAndSucceed(
		fmt.Sprintf(`/var/vcap/packages/postgres-9.4/bin/psql -U vcap "%s" --command="%s"`, database, command),
	)
}

func runOnPostgresVMAndSucceed(command string) *gexec.Session {
	session := RunOnInstance(
		"postgres-dev", "postgres", "0", command,
	)
	Expect(session).To(gexec.Exit(0))

	return session
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

func getUploadCommand(localPath, remotePath, instanceName, instanceIndex string) string {
	return fmt.Sprintf(
		"scp --gw-user=%s --gw-host=%s --gw-private-key=%s %s %s/%s:%s",
		MustHaveEnv("BOSH_GATEWAY_USER"),
		MustHaveEnv("BOSH_GATEWAY_HOST"),
		MustHaveEnv("BOSH_GATEWAY_KEY"),
		localPath,
		instanceName,
		instanceIndex,
		remotePath,
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
