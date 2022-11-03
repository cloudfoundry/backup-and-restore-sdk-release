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
	It("fails when neither backup nor restore flag is set", func() {
		command := exec.Command(binaryPath)
		fmt.Fprintf(GinkgoWriter, "Running command: %s\n", command.String())
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session).Should(gexec.Exit())
		Expect(session.ExitCode()).NotTo(Equal(0))
		Expect(session.Err).To(gbytes.Say("missing --backup or --restore flag"))
	})

	It("fails when both backup nor restore flags are provided", func() {
		command := exec.Command(binaryPath, "--backup", "--restore")
		fmt.Fprintf(GinkgoWriter, "Running command: %s\n", command.String())
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())

		Eventually(session).Should(gexec.Exit())
		Expect(session.ExitCode()).NotTo(Equal(0))
		Expect(session.Err).To(gbytes.Say("only one of: --backup or --restore can be provided"))
	})
})
