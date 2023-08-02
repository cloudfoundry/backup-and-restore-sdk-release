package main_test

import (
	"fmt"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Main", func() {
	It("it fails when the --config flag is not present", func() {
		session, err := gexec.Start(exec.Command(binaryPath, ""), GinkgoWriter, GinkgoWriter)

		exitsWithErrorMsg(err, session, "missing --config flag")
	})

	It("it fails when no action flag is provided", func() {
		session, err := gexec.Start(exec.Command(binaryPath, "--config", "a-config-path"), GinkgoWriter, GinkgoWriter)

		exitsWithErrorMsg(err, session, "exactly one action flag must be provided")
	})

	Context("when multiple actions are provided", func() {
		for _, actions := range [][]string{
			{"versioned-backup", "versioned-restore"},
			{"versioned-backup", "unversioned-backup-start"},
			{"versioned-backup", "unversioned-backup-complete"},
			{"versioned-backup", "unversioned-restore"},
		} {
			It(fmt.Sprintf("fails with --%s and --%s", actions[0], actions[1]), func() {
				session, err := gexec.Start(
					exec.Command(binaryPath, fmt.Sprintf("--%s", actions[0]), fmt.Sprintf("--%s", actions[1]), "--config", "a-config-path"),
					GinkgoWriter,
					GinkgoWriter,
				)
				exitsWithErrorMsg(err, session, "exactly one action flag must be provided")
			})
		}
	})

	Context("when the action requires the artifact flag", func() {
		for _, action := range []string{"versioned-backup", "versioned-restore", "unversioned-backup-start", "unversioned-restore"} {
			It(fmt.Sprintf("fails when the --artifact flag is missing for --%s", action), func() {
				session, err := gexec.Start(
					exec.Command(binaryPath, fmt.Sprintf("--%s", action), "--config", "a-config-path"),
					GinkgoWriter,
					GinkgoWriter,
				)

				exitsWithErrorMsg(err, session, "missing --artifact flag")
			})
		}
	})

	Context("when the action requires the existing-artifact flag", func() {
		It(fmt.Sprintf("fails when the --existing-artifact is missing for --unversioned-backup-start"), func() {
			session, err := gexec.Start(
				exec.Command(binaryPath, "--unversioned-backup-start", "--config", "a-config-path", "--artifact", "a-artifact-path"),
				GinkgoWriter,
				GinkgoWriter,
			)
			exitsWithErrorMsg(err, session, "missing --existing-artifact flag")
		})

		It(fmt.Sprintf("fails when the --existing-artifact is missing for --unversioned-backup-complete"), func() {
			session, err := gexec.Start(
				exec.Command(binaryPath, "--unversioned-backup-complete", "--config", "a-config-path"),
				GinkgoWriter,
				GinkgoWriter,
			)
			exitsWithErrorMsg(err, session, "missing --existing-artifact flag")
		})
	})
})

func exitsWithErrorMsg(err error, session *gexec.Session, errMsg string) {
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gexec.Exit())
	Expect(session.ExitCode()).NotTo(Equal(0))
	Expect(session.Err).To(gbytes.Say(errMsg))
}
