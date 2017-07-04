package database_backup_and_restore

import (
	"fmt"
	"os/exec"
	"time"

	"io/ioutil"

	"os"

	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/pivotal-cf-experimental/go-binmock"
)

type GeneratorFunction func() (string, error)
type TestEntry struct {
	arguments       string
	expectedOutput  string
	configGenerator GeneratorFunction
}

var _ = Describe("Backup and Restore DB Utility", func() {
	var fakePgDump *binmock.Mock
	var fakePgRestore *binmock.Mock
	var session *gexec.Session
	var username = "testuser"
	var host = "127.0.0.1"
	var port = 1234
	var databaseName = "mycooldb"
	var password = "password"
	var adapter = "postgres"
	var artifactFile string
	var path string
	var err error

	BeforeEach(func() {
		path, err = gexec.Build("github.com/pivotal-cf/database-backup-and-restore/cmd/database-backup-restore")
		Expect(err).NotTo(HaveOccurred())
	})

	Context("incorrect cli usage", func() {
		var entriesToTest []TableEntry

		entriesToTest = []TableEntry{
			Entry("two actions provided", TestEntry{
				arguments:      "--backup --restore",
				expectedOutput: "Only one of: --backup or --restore can be provided",
			}),
			Entry("no action provided", TestEntry{
				arguments:      "--artifact-file /foo --config foo",
				expectedOutput: "Missing --backup or --restore flag",
			}),
			Entry("no config is passed", TestEntry{
				arguments:      "--backup --artifact-file /foo",
				expectedOutput: "Missing --config flag",
			}),
			Entry("the config is not accessible", TestEntry{
				arguments:      "--backup --artifact-file /foo --config /foo/bar/bar.json",
				expectedOutput: "no such file",
			}),
			Entry("is not a valid json", TestEntry{
				arguments:       "--backup --artifact-file /foo --config %s",
				configGenerator: invalidConfig,
				expectedOutput:  "Could not parse config json",
			}),
			Entry("unsupported adapter", TestEntry{
				arguments:       "--backup --artifact-file /foo --config %s",
				configGenerator: invalidAdapterConfig,
				expectedOutput:  "Unsupported adapter foo-server",
			}),
			Entry("the artifact-file is not provided", TestEntry{
				arguments:       "--backup --config %s",
				configGenerator: validConfig,
				expectedOutput:  "Missing --artifact-file flag",
			}),
			Entry("PG_DUMP_PATH is not set", TestEntry{
				arguments:       "--backup --artifact-file /foo --config %s",
				configGenerator: validConfig,
				expectedOutput:  "PG_DUMP_PATH must be set",
			}),
			Entry("PG_RESTORE_PATH is not set", TestEntry{
				arguments:       "--restore --artifact-file /foo --config %s",
				configGenerator: validConfig,
				expectedOutput:  "PG_RESTORE_PATH must be set",
			}),
		}

		DescribeTable("raises the appropriate error when",
			func(entry TestEntry) {
				if entry.configGenerator != nil {
					configPath, err := entry.configGenerator()
					Expect(err).NotTo(HaveOccurred())
					entry.arguments = fmt.Sprintf(entry.arguments, configPath)
					defer os.Remove(configPath)
				}
				args := strings.Split(entry.arguments, " ")
				cmd := exec.Command(path, args...)

				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
				Eventually(session).Should(gexec.Exit(1))
				Expect(session.Err).To(gbytes.Say(entry.expectedOutput))
			},
			entriesToTest...,
		)
	})

	Context("--backup", func() {
		var cmdActionFlag string
		BeforeEach(func() {
			cmdActionFlag = "--backup"
		})

		var configFile *os.File

		BeforeEach(func() {
			fakePgDump = binmock.NewBinMock("pg_dump")
			fakePgDump.WhenCalled().WillExitWith(0)

			artifactFile = tempFilePath()
			configFile, err = ioutil.TempFile(os.TempDir(), time.Now().String())
			Expect(err).NotTo(HaveOccurred())
			fmt.Fprintf(
				configFile,
				`{"username":"%s","password":"%s","host":"%s","port":%d,"database":"%s","adapter":"%s"}`,
				username,
				password,
				host,
				port,
				databaseName,
				adapter,
			)
		})

		JustBeforeEach(func() {
			cmd := exec.Command(path, "--artifact-file", artifactFile, "--config", configFile.Name(), cmdActionFlag)
			cmd.Env = append(cmd.Env, fmt.Sprintf("PG_DUMP_PATH=%s", fakePgDump.Path))

			session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(session).Should(gexec.Exit())
		})

		AfterEach(func() {
			os.Remove(artifactFile)
		})

		It("calls pg_dump with the correct arguments", func() {
			expectedArgs := []string{
				"-v",
				fmt.Sprintf("--user=%s", username),
				fmt.Sprintf("--host=%s", host),
				fmt.Sprintf("--port=%d", port),
				"--format=custom",
				fmt.Sprintf("--file=%s", artifactFile),
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

		Context("and pg_dump fails", func() {
			BeforeEach(func() {
				fakePgDump = binmock.NewBinMock("pg_dump")
				fakePgDump.WhenCalled().WillExitWith(1)
			})

			It("also fails", func() {
				Eventually(session).Should(gexec.Exit(1))
			})
		})
	})

	Context("--restore", func() {
		var configFile *os.File
		var cmdActionFlag string

		BeforeEach(func() {
			cmdActionFlag = "--restore"

			fakePgRestore = binmock.NewBinMock("pg_restore")
			fakePgRestore.WhenCalled().WillExitWith(0)

			artifactFile = tempFilePath()
			configFile, err = ioutil.TempFile(os.TempDir(), time.Now().String())
			Expect(err).NotTo(HaveOccurred())
			fmt.Fprintf(
				configFile,
				`{"username":"%s","password":"%s","host":"%s","port":%d,"database":"%s","adapter":"%s"}`,
				username,
				password,
				host,
				port,
				databaseName,
				adapter,
			)

		})

		JustBeforeEach(func() {
			cmd := exec.Command(path, "--artifact-file", artifactFile, "--config", configFile.Name(), cmdActionFlag)
			cmd.Env = append(cmd.Env, fmt.Sprintf("PG_RESTORE_PATH=%s", fakePgRestore.Path))

			session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(session).Should(gexec.Exit())
		})

		AfterEach(func() {
			os.Remove(artifactFile)
		})

		It("calls pg_restore with the correct arguments", func() {
			expectedArgs := []string{
				"-v",
				fmt.Sprintf("--user=%s", username),
				fmt.Sprintf("--host=%s", host),
				fmt.Sprintf("--port=%d", port),
				"--format=custom",
				fmt.Sprintf("--dbname=%s", databaseName),
				"--clean",
				artifactFile,
			}

			Expect(fakePgRestore.Invocations()).To(HaveLen(1))
			Expect(fakePgRestore.Invocations()[0].Args()).Should(ConsistOf(expectedArgs))
			Expect(fakePgRestore.Invocations()[0].Env()).Should(HaveKeyWithValue("PGPASSWORD", password))
		})

		It("succeeds", func() {
			Expect(session).Should(gexec.Exit(0))
		})

		Context("and pg_restore fails", func() {
			BeforeEach(func() {
				fakePgRestore = binmock.NewBinMock("pg_restore")
				fakePgRestore.WhenCalled().WillExitWith(1)
			})

			It("also fails", func() {
				Eventually(session).Should(gexec.Exit(1))
			})
		})
	})
})

func tempFilePath() string {
	tmpfile, err := ioutil.TempFile("", "")
	Expect(err).NotTo(HaveOccurred())
	tmpfile.Close()
	return tmpfile.Name()
}

func invalidConfig() (string, error) {
	invalidJsonConfig, err := ioutil.TempFile(os.TempDir(), "")
	if err != nil {
		return "", err
	}
	fmt.Fprintf(invalidJsonConfig, "foo!")
	return invalidJsonConfig.Name(), nil
}

func invalidAdapterConfig() (string, error) {
	invalidAdapterConfig, err := ioutil.TempFile(os.TempDir(), "")
	if err != nil {
		return "", err
	}
	fmt.Fprintf(invalidAdapterConfig, `{"adapter":"foo-server"}`)
	return invalidAdapterConfig.Name(), nil
}

func validConfig() (string, error) {
	validConfig, err := ioutil.TempFile(os.TempDir(), "")
	if err != nil {
		return "", err
	}

	fmt.Fprintf(validConfig,
		`{"username":"testuser","password":"password","host":"127.0.0.1","port":1234,"database":"mycooldb","adapter":"postgres"}`,
	)
	return validConfig.Name(), nil
}
