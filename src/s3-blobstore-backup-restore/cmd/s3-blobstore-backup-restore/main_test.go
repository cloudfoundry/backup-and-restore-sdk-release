package main_test

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"os/exec"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Main", func() {
	It("it fails when the --config flag is not present", func() {
		session := runCommand(binaryPath, "")
		exitsWithErrorMsg(session, "missing --config flag")
	})

	It("it fails when no action flag is provided", func() {
		session := runCommand(binaryPath, "--config", "a-config-path")
		exitsWithErrorMsg(session, "exactly one action flag must be provided")
	})

	Context("when multiple actions are provided", func() {
		for _, actions := range [][]string{
			{"versioned-backup", "versioned-restore"},
			{"versioned-backup", "unversioned-backup-start"},
			{"versioned-backup", "unversioned-backup-complete"},
			{"versioned-backup", "unversioned-restore"},
		} {
			It(fmt.Sprintf("fails with --%s and --%s", actions[0], actions[1]), func() {
				session := runCommand(binaryPath, fmt.Sprintf("--%s", actions[0]), fmt.Sprintf("--%s", actions[1]), "--config", "a-config-path")
				exitsWithErrorMsg(session, "exactly one action flag must be provided")
			})
		}
	})

	Context("when the action requires the artifact flag", func() {
		for _, action := range []string{"versioned-backup", "versioned-restore", "unversioned-backup-start", "unversioned-restore"} {
			It(fmt.Sprintf("fails when the --artifact flag is missing for --%s", action), func() {
				session := runCommand(binaryPath, fmt.Sprintf("--%s", action), "--config", "a-config-path")
				exitsWithErrorMsg(session, "missing --artifact flag")
			})
		}
	})

	Context("when the action requires the existing-artifact flag", func() {
		It(fmt.Sprintf("fails when the --existing-artifact is missing for --unversioned-backup-start"), func() {
			session := runCommand(binaryPath, "--unversioned-backup-start", "--config", "a-config-path", "--artifact", "a-artifact-path")
			exitsWithErrorMsg(session, "missing --existing-artifact flag")
		})

		It(fmt.Sprintf("fails when the --existing-artifact is missing for --unversioned-backup-complete"), func() {
			session := runCommand(binaryPath, "--unversioned-backup-complete", "--config", "a-config-path")
			exitsWithErrorMsg(session, "missing --existing-artifact flag")
		})
	})
})

func runCommand(path string, args ...string) *gexec.Session {
	command := exec.Command(path, args...)
	fmt.Fprintf(GinkgoWriter, "Running command: %s\n", command.String())
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	return session
}

func exitsWithErrorMsg(session *gexec.Session, errMsg string) {
	Eventually(session).Should(gexec.Exit())
	Expect(session.ExitCode()).NotTo(Equal(0))
	Expect(session.Err).To(gbytes.Say(errMsg))
}
