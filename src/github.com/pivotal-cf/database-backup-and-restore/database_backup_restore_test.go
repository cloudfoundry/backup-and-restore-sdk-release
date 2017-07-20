package database_backup_and_restore

import (
	"fmt"
	"os/exec"
	"time"

	"io/ioutil"

	"os"

	"strings"

	"log"

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
	var fakeDump *binmock.Mock
	var fakeRestore *binmock.Mock
	var fakeClient *binmock.Mock
	var session *gexec.Session
	var username = "testuser"
	var host = "127.0.0.1"
	var port = 1234
	var databaseName = "mycooldb"
	var password = "password"
	var adapter string
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
				configGenerator: validPgConfig,
				expectedOutput:  "Missing --artifact-file flag",
			}),
			Entry("PG_DUMP_PATH is not set", TestEntry{
				arguments:       "--backup --artifact-file /foo --config %s",
				configGenerator: validPgConfig,
				expectedOutput:  "PG_DUMP_PATH must be set",
			}),
			Entry("PG_RESTORE_PATH is not set", TestEntry{
				arguments:       "--restore --artifact-file /foo --config %s",
				configGenerator: validPgConfig,
				expectedOutput:  "PG_RESTORE_PATH must be set",
			}),
			Entry("MYSQL_DUMP_PATH is not set", TestEntry{
				arguments:       "--backup --artifact-file /foo --config %s",
				configGenerator: validMysqlConfig,
				expectedOutput:  "MYSQL_DUMP_PATH must be set",
			}),
			Entry("MYSQL_CLIENT_PATH is not set", TestEntry{
				arguments:       "--restore --artifact-file /foo --config %s",
				configGenerator: validMysqlConfig,
				expectedOutput:  "MYSQL_CLIENT_PATH must be set",
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
		var configFile *os.File
		var cmdActionFlag = "--backup"

		JustBeforeEach(func() {
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

		AfterEach(func() {
			os.Remove(artifactFile)
		})

		Context("mysql", func() {
			BeforeEach(func() {
				adapter = "mysql"
				fakeDump = binmock.NewBinMock("mysqldump")
				fakeDump.WhenCalledWith("-V").WillPrintToStdOut("mysqldump  Ver 10.16 Distrib 10.1.24-MariaDB, for Linux (x86_64)")
				fakeDump.WhenCalled().WillExitWith(0)

				fakeClient = binmock.NewBinMock("mysql")
				// fakeClient.WhenCalled().WillPrintToStdOut("10.1.24-MariaDB-wsrep")
				fakeClient.WhenCalledWith("--skip-column-names",
					"--silent",
					fmt.Sprintf("--user=%s", username),
					fmt.Sprintf("--password=%s", password),
					fmt.Sprintf("--host=%s", host),
					fmt.Sprintf("--port=%d", port),
					`--execute=SELECT VERSION()`).WillPrintToStdOut("10.1.24-MariaDB-wsrep")
			})

			JustBeforeEach(func() {
				cmd := exec.Command(path, "--artifact-file", artifactFile, "--config", configFile.Name(), cmdActionFlag)
				cmd.Env = append(cmd.Env, fmt.Sprintf("MYSQL_DUMP_PATH=%s", fakeDump.Path))
				cmd.Env = append(cmd.Env, fmt.Sprintf("MYSQL_CLIENT_PATH=%s", fakeClient.Path))

				session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
				Eventually(session).Should(gexec.Exit())
			})

			It("calls mysqldump with the correct arguments", func() {
				Expect(fakeDump.Invocations()).To(HaveLen(2))

				By("first checking the version", func() {
					Expect(fakeDump.Invocations()[0].Args()).Should(ConsistOf("-V"))
				})

				By("then calling dump", func() {
					expectedArgs := []string{
						"-v",
						"--single-transaction",
						"--skip-add-locks",
						fmt.Sprintf("--user=%s", username),
						fmt.Sprintf("--host=%s", host),
						fmt.Sprintf("--port=%d", port),
						fmt.Sprintf("--result-file=%s", artifactFile),
						databaseName,
					}

					Expect(fakeDump.Invocations()[1].Args()).Should(ConsistOf(expectedArgs))
					Expect(fakeDump.Invocations()[1].Env()).Should(HaveKeyWithValue("MYSQL_PWD", password))
				})
			})

			It("succeeds", func() {
				Expect(session).Should(gexec.Exit(0))
			})

			Context("and mysqldump fails", func() {
				BeforeEach(func() {
					fakeDump = binmock.NewBinMock("mysqldump")
					fakeDump.WhenCalled().WillExitWith(1)
				})

				It("also fails", func() {
					Eventually(session).Should(gexec.Exit(1))
				})
			})

			Context("mysqldump has a different major version than the server", func() {
				BeforeEach(func() {
					fakeClient = binmock.NewBinMock("mysql")
					fakeClient.WhenCalledWith(
						"--skip-column-names",
						"--silent",
						fmt.Sprintf("--user=%s", username),
						fmt.Sprintf("--password=%s", password),
						fmt.Sprintf("--host=%s", host),
						fmt.Sprintf("--port=%d", port),
						`--execute=SELECT VERSION()`).WillPrintToStdOut("9.1.24-MariaDB-wsrep")
				})

				It("fails because of a version mismatch", func() {
					Expect(session).Should(gexec.Exit(1))
					Expect(string(session.Err.Contents())).Should(ContainSubstring("major/minor version mismatch"))
				})
			})

			Context("mysqldump has a different minor version than the server", func() {
				BeforeEach(func() {
					fakeClient = binmock.NewBinMock("mysql")
					fakeClient.WhenCalledWith(
						"--skip-column-names",
						"--silent",
						fmt.Sprintf("--user=%s", username),
						fmt.Sprintf("--password=%s", password),
						fmt.Sprintf("--host=%s", host),
						fmt.Sprintf("--port=%d", port),
						`--execute=SELECT VERSION()`).WillPrintToStdOut("10.0.24-MariaDB-wsrep")
				})

				It("fails because of a version mismatch", func() {
					Expect(session).Should(gexec.Exit(1))
					Expect(string(session.Err.Contents())).Should(ContainSubstring("major/minor version mismatch"))
				})
			})

			Context("mysqldump has a different patch version than the server", func() {
				BeforeEach(func() {
					fakeClient = binmock.NewBinMock("mysql")
					fakeClient.WhenCalledWith(
						"--skip-column-names",
						"--silent",
						fmt.Sprintf("--user=%s", username),
						fmt.Sprintf("--password=%s", password),
						fmt.Sprintf("--host=%s", host),
						fmt.Sprintf("--port=%d", port),
						`--execute=SELECT VERSION()`).WillPrintToStdOut("10.1.22-MariaDB-wsrep")
				})

				It("succeeds despite different patch versions", func() {
					Expect(session).Should(gexec.Exit(0))
				})
			})
		})

		Context("postgres", func() {
			BeforeEach(func() {
				adapter = "postgres"
				fakeDump = binmock.NewBinMock("pg_dump")
				fakeDump.WhenCalled().WillExitWith(0)
			})

			JustBeforeEach(func() {
				cmd := exec.Command(path, "--artifact-file", artifactFile, "--config", configFile.Name(), cmdActionFlag)
				cmd.Env = append(cmd.Env, fmt.Sprintf("PG_DUMP_PATH=%s", fakeDump.Path))

				session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
				Eventually(session).Should(gexec.Exit())
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

				Expect(fakeDump.Invocations()).To(HaveLen(1))
				Expect(fakeDump.Invocations()[0].Args()).Should(ConsistOf(expectedArgs))
				Expect(fakeDump.Invocations()[0].Env()).Should(HaveKeyWithValue("PGPASSWORD", password))
			})

			It("calls pg_dump with the correct env vars", func() {
				Expect(fakeDump.Invocations()).To(HaveLen(1))
				Expect(fakeDump.Invocations()[0].Env()).Should(HaveKeyWithValue("PGPASSWORD", password))
			})

			It("succeeds", func() {
				Expect(session).Should(gexec.Exit(0))
			})

			Context("and pg_dump fails", func() {
				BeforeEach(func() {
					fakeDump = binmock.NewBinMock("pg_dump")
					fakeDump.WhenCalled().WillExitWith(1)
				})

				It("also fails", func() {
					Eventually(session).Should(gexec.Exit(1))
				})
			})
		})
	})

	Context("--restore", func() {
		var configFile *os.File
		var cmdActionFlag = "--restore"

		JustBeforeEach(func() {
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

		AfterEach(func() {
			os.Remove(artifactFile)
		})

		Context("mysql", func() {
			BeforeEach(func() {
				adapter = "mysql"
				fakeRestore = binmock.NewBinMock("mysql")
				fakeRestore.WhenCalled().WillExitWith(0)
			})

			JustBeforeEach(func() {
				err := ioutil.WriteFile(artifactFile, []byte("SOME BACKUP SQL"), 0644)
				if err != nil {
					log.Fatalln("Failed to write to artifact file, %s", err)
				}

				cmd := exec.Command(path, "--artifact-file", artifactFile, "--config", configFile.Name(), cmdActionFlag)
				cmd.Env = append(cmd.Env, fmt.Sprintf("MYSQL_CLIENT_PATH=%s", fakeRestore.Path))

				session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
				Eventually(session).Should(gexec.Exit())
			})

			It("calls mysql with the correct arguments", func() {
				expectedArgs := []string{
					"-v",
					fmt.Sprintf("--user=%s", username),
					fmt.Sprintf("--host=%s", host),
					fmt.Sprintf("--port=%d", port),
					databaseName,
				}

				Expect(fakeRestore.Invocations()).To(HaveLen(1))
				Expect(fakeRestore.Invocations()[0].Args()).Should(ConsistOf(expectedArgs))
				Expect(fakeRestore.Invocations()[0].Stdin()).Should(ConsistOf("SOME BACKUP SQL"))
				Expect(fakeRestore.Invocations()[0].Env()).Should(HaveKeyWithValue("MYSQL_PWD", password))
			})

			It("succeeds", func() {
				Expect(session).Should(gexec.Exit(0))
			})

			Context("and mysql fails", func() {
				BeforeEach(func() {
					fakeRestore = binmock.NewBinMock("mysql")
					fakeRestore.WhenCalled().WillExitWith(1)
				})

				It("also fails", func() {
					Eventually(session).Should(gexec.Exit(1))
				})
			})
		})

		Context("postgres", func() {
			BeforeEach(func() {
				adapter = "postgres"
				fakeRestore = binmock.NewBinMock("pg_restore")
				fakeRestore.WhenCalled().WillExitWith(0)
			})

			JustBeforeEach(func() {
				cmd := exec.Command(path, "--artifact-file", artifactFile, "--config", configFile.Name(), cmdActionFlag)
				cmd.Env = append(cmd.Env, fmt.Sprintf("PG_RESTORE_PATH=%s", fakeRestore.Path))

				session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
				Eventually(session).Should(gexec.Exit())
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

				Expect(fakeRestore.Invocations()).To(HaveLen(1))
				Expect(fakeRestore.Invocations()[0].Args()).Should(ConsistOf(expectedArgs))
				Expect(fakeRestore.Invocations()[0].Env()).Should(HaveKeyWithValue("PGPASSWORD", password))
			})

			It("succeeds", func() {
				Expect(session).Should(gexec.Exit(0))
			})

			Context("and pg_restore fails", func() {
				BeforeEach(func() {
					fakeRestore = binmock.NewBinMock("pg_restore")
					fakeRestore.WhenCalled().WillExitWith(1)
				})

				It("also fails", func() {
					Eventually(session).Should(gexec.Exit(1))
				})
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

func validMysqlConfig() (string, error) {
	validConfig, err := ioutil.TempFile(os.TempDir(), "")
	if err != nil {
		return "", err
	}

	fmt.Fprintf(validConfig,
		`{"username":"testuser","password":"password","host":"127.0.0.1","port":1234,"database":"mycooldb","adapter":"mysql"}`,
	)
	return validConfig.Name(), nil
}

func validPgConfig() (string, error) {
	validConfig, err := ioutil.TempFile(os.TempDir(), "")
	if err != nil {
		return "", err
	}

	fmt.Fprintf(validConfig,
		`{"username":"testuser","password":"password","host":"127.0.0.1","port":1234,"database":"mycooldb","adapter":"postgres"}`,
	)
	return validConfig.Name(), nil
}
