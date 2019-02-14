package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"os/exec"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Main", func() {
	It("it fails when there are no flags provided", func() {
		session, err := gexec.Start(exec.Command(binaryPath), GinkgoWriter, GinkgoWriter)

		exitsWithErrorMsg(err, session, "missing --backup or --restore flag")
	})

	It("it fails when the --config flag is not present", func() {
		session, err := gexec.Start(exec.Command(binaryPath, "--restore"), GinkgoWriter, GinkgoWriter)

		exitsWithErrorMsg(err, session, "missing --config flag")
	})

	It("it fails when the --artifact-file flag is not present", func() {
		session, err := gexec.Start(exec.Command(binaryPath, "--restore", "--config", "a-config-path"), GinkgoWriter, GinkgoWriter)

		exitsWithErrorMsg(err, session, "missing --artifact-file")
	})

	Context("when --backup flag is passed", func() {
		var args []string

		BeforeEach(func() {
			args = []string{"--backup", "--config", "a-config-path", "--artifact-file", "an-artifact-file"}
		})

		Context("and --unversioned-backup-completer flags is passed", func() {
			It("it fails when the --existing-backup-blobs-artifact flag is not present", func() {
				args = append(args, "--unversioned-backup-completer")

				session, err := gexec.Start(exec.Command(binaryPath, args...), GinkgoWriter, GinkgoWriter)

				exitsWithErrorMsg(err, session, "missing --existing-backup-blobs-artifact")
			})
		})

		Context("and --unversioned-backup-starter flags is passed", func() {
			It("it fails when the --existing-backup-blobs-artifact flag is not present", func() {
				args = append(args, "--unversioned-backup-starter")

				session, err := gexec.Start(exec.Command(binaryPath, args...), GinkgoWriter, GinkgoWriter)

				exitsWithErrorMsg(err, session, "missing --existing-backup-blobs-artifact")
			})
		})

		Context("and both --unversioned-backup-starter and --unversioned-backup-completer flags are passed", func() {
			It("errors", func() {
				args = append(args, "--unversioned-backup-starter", "--unversioned-backup-completer")

				session, err := gexec.Start(exec.Command(binaryPath, args...), GinkgoWriter, GinkgoWriter)

				exitsWithErrorMsg(err, session, "at most one of: --unversioned-backup-starter or --unversioned-backup-completer can be provided")
			})
		})
	})
})

func exitsWithErrorMsg(err error, session *gexec.Session, errMsg string) {
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gexec.Exit())
	Expect(session.ExitCode()).NotTo(Equal(0))
	Expect(session.Err).To(gbytes.Say(errMsg))
}
