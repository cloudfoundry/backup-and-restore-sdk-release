package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"os/exec"

	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Main", func() {
	It("fails when neither backup nor restore flag is set", func() {
		session, err := gexec.Start(exec.Command(binaryPath), GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session).Should(gexec.Exit())
		Expect(session.ExitCode()).NotTo(Equal(0))
		Expect(session.Err).To(gbytes.Say("missing --backup or --restore flag"))
	})

	It("fails when both backup nor restore flags are provided", func() {
		session, err := gexec.Start(exec.Command(binaryPath, "--backup", "--restore"), GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session).Should(gexec.Exit())
		Expect(session.ExitCode()).NotTo(Equal(0))
		Expect(session.Err).To(gbytes.Say("only one of: --backup or --restore can be provided"))
	})
})
