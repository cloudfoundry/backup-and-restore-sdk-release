package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"os/exec"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Main", func() {
	It("fails when neither backup, restore or unlock flag is set", func() {
		session, err := gexec.Start(exec.Command(binaryPath), GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session).Should(gexec.Exit())
		Expect(session.ExitCode()).NotTo(Equal(0))
		Expect(session.Err).To(gbytes.Say("missing --unlock, --backup or --restore flag"))
	})

	It("fails when both backup, restore and unlock flags are provided", func() {
		session, err := gexec.Start(exec.Command(binaryPath, "--backup", "--restore"), GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session).Should(gexec.Exit())
		Expect(session.ExitCode()).NotTo(Equal(0))
		Expect(session.Err).To(gbytes.Say("only one of: --unlock, --backup or --restore can be provided"))

		session, err = gexec.Start(exec.Command(binaryPath, "--unlock", "--restore"), GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session).Should(gexec.Exit())
		Expect(session.ExitCode()).NotTo(Equal(0))
		Expect(session.Err).To(gbytes.Say("only one of: --unlock, --backup or --restore can be provided"))

		session, err = gexec.Start(exec.Command(binaryPath, "--backup", "--unlock"), GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session).Should(gexec.Exit())
		Expect(session.ExitCode()).NotTo(Equal(0))
		Expect(session.Err).To(gbytes.Say("only one of: --unlock, --backup or --restore can be provided"))
	})
})
