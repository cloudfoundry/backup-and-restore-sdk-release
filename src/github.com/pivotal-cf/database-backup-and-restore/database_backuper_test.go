package database_backup_and_restore

import (
	"fmt"
	"os/exec"
	"time"

	"io/ioutil"

	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
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
	var adapter = "postgres"
	var outputFile string
	var path string
	var err error

	BeforeEach(func() {
		path, err = gexec.Build("github.com/pivotal-cf/database-backup-and-restore/cmd/database-backuper")
		Expect(err).NotTo(HaveOccurred())
	})

	Context("when the arguments are wrong", func() {
		It("exit with error if no config is passed", func() {
			cmd = exec.Command(path)

			session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(session).Should(gexec.Exit(1))
			Expect(session.Err).To(gbytes.Say(`missing argument: config.json`))
		})

		It("exit with error if the config is not accessible", func() {
			cmd = exec.Command(path, "/foo/bar/bar.json")

			session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(session).Should(gexec.Exit(1))
			Expect(session.Err).To(gbytes.Say(`no such file`))
		})

		It("exit with error if the config is not a valid json", func() {
			configFile, err := ioutil.TempFile(os.TempDir(), time.Now().String())
			Expect(err).NotTo(HaveOccurred())
			fmt.Fprintf(configFile, "foo!")
			cmd = exec.Command(path, configFile.Name())

			session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(session).Should(gexec.Exit(1))
			Expect(session.Err).To(gbytes.Say(`Could not parse config json`))
		})

		It("exit with error if unsupported adapter", func() {
			outputFile = tempFilePath()
			configFile, err := ioutil.TempFile(os.TempDir(), time.Now().String())
			Expect(err).NotTo(HaveOccurred())
			fmt.Fprintf(configFile, `{"adapter":"foo-server"}`)

			cmd = exec.Command(path, configFile.Name())

			session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(session).Should(gexec.Exit(1))
			Expect(session.Err).To(gbytes.Say(`Unsupported adapter foo-server`))
		})
	})

	Context("when the config.json is provided", func() {
		BeforeEach(func() {
			fakePgDump = binmock.NewBinMock("pg_dump")
			fakePgDump.WhenCalled().WillExitWith(0)

			outputFile = tempFilePath()
			configFile, err := ioutil.TempFile(os.TempDir(), time.Now().String())
			Expect(err).NotTo(HaveOccurred())
			fmt.Fprintf(
				configFile,
				`{"username":"%s","password":"%s","host":"%s","port":"%s","database":"%s","adapter":"%s","output_file":"%s"}`,
				username,
				password,
				host,
				port,
				databaseName,
				adapter,
				outputFile,
			)

			cmd = exec.Command(path, configFile.Name())
			cmd.Env = append(cmd.Env, fmt.Sprintf("PG_DUMP_PATH=%s", fakePgDump.Path))

			session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(session).Should(gexec.Exit())
		})

		AfterEach(func() {
			os.Remove(outputFile)
		})

		It("calls pg_dump with the correct arguments", func() {
			expectedArgs := []string{
				fmt.Sprintf("--user=%s", username),
				fmt.Sprintf("--host=%s", host),
				fmt.Sprintf("--port=%s", port),
				"--format=custom",
				fmt.Sprintf("--file=%s", outputFile),
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
	})
})

func tempFilePath() string {
	tmpfile, err := ioutil.TempFile("", "")
	Expect(err).NotTo(HaveOccurred())
	tmpfile.Close()
	return tmpfile.Name()
}
