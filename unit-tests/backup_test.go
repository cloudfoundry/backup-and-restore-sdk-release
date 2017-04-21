package unit_tests

import (
	"fmt"
	"os/exec"

	"io/ioutil"

	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/pivotal-cf-experimental/go-binmock"
)

var _ = Describe("Backup", func() {
	var fakePgDump *binmock.Mock
	var cmd *exec.Cmd
	var session *gexec.Session
	var username = "testuser"
	var host = "127.0.0.1"
	var port = "1234"
	var databaseName = "mycooldb"
	var password = "password"
	var outputFile string
	AfterEach(func() {
		os.Remove(outputFile)
	})

	BeforeEach(func() {
		fakePgDump = binmock.NewBinMock("pg_dump")
		fakePgDump.WhenCalled().WillExitWith(0).WillPrintToStdOut("pg_dump_output")

		outputFile = tempFilePath()

		cmd = exec.Command("../jobs/database_backuper/templates/backup")
		cmd.Env = []string{
			fmt.Sprintf("PG_DUMP_PATH=%s", fakePgDump.Path),
			fmt.Sprintf("USER=%s", username),
			fmt.Sprintf("HOST=%s", host),
			fmt.Sprintf("PORT=%s", port),
			fmt.Sprintf("DATABASE=%s", databaseName),
			fmt.Sprintf("PASSWORD=%s", password),
			fmt.Sprintf("OUTPUT_FILE=%s", outputFile),
		}

		var err error
		session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
		Expect(err).ToNot(HaveOccurred())
		Eventually(session).Should(gexec.Exit())
	})

	It("calls pg_dump with the correct arguments", func() {
		expectedArgs := []string{
			fmt.Sprintf("--user=%s", username),
			fmt.Sprintf("--host=%s", host),
			fmt.Sprintf("--port=%s", port),
			"--format=custom",
			databaseName,
		}

		Expect(fakePgDump.Invocations()).To(HaveLen(1))
		Expect(fakePgDump.Invocations()[0].Args()).Should(ConsistOf(expectedArgs))
		Expect(fakePgDump.Invocations()[0].Env()).Should(HaveKeyWithValue("PGPASSWORD", password))
	})

	It("calls pg_dump with the correct env vars", func() {
		Expect(fakePgDump.Invocations()).To(HaveLen(1))
		Expect(fakePgDump.Invocations()[0].Env()).Should(HaveKeyWithValue("PGPASSWORD", password))
	})

	It("succeeds", func() {
		Expect(session).Should(gexec.Exit(0))
	})

	It("pushes the pg_dump output to the output file", func() {
		Expect(ioutil.ReadFile(outputFile)).To(ContainSubstring("pg_dump_output"))
	})
})

func tempFilePath() string {
	tmpfile, err := ioutil.TempFile("", "")
	Expect(err).NotTo(HaveOccurred())
	tmpfile.Close()
	return tmpfile.Name()
}
