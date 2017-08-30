// Copyright (C) 2017-Present Pivotal Software, Inc. All rights reserved.
//
// This program and the accompanying materials are made available under
// the terms of the under the Apache License, Version 2.0 (the "License”);
// you may not use this file except in compliance with the License.
//
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//
// See the License for the specific language governing permissions and
// limitations under the License.

package database_backup_and_restore

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/pivotal-cf-experimental/go-binmock"
)

type TestEntry struct {
	arguments       string
	expectedOutput  string
	configGenerator func() (string, error)
}

var _ = Describe("Backup and Restore DB Utility", func() {
	var fakeDump *binmock.Mock
	var fakeDumpPostgres96 *binmock.Mock
	var fakeRestore *binmock.Mock
	var fakeClient *binmock.Mock
	var session *gexec.Session
	var username = "testuser"
	var host = "127.0.0.1"
	var port = 1234
	var databaseName = "mycooldb"
	var password = "password"
	var artifactFile string
	var path string
	var err error
	var configFile *os.File

	BeforeEach(func() {
		path, err = gexec.Build("github.com/pivotal-cf/database-backup-and-restore/cmd/database-backup-restore")
		Expect(err).NotTo(HaveOccurred())

		artifactFile = tempFilePath()
	})

	AfterEach(func() {
		os.Remove(artifactFile)
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
			XEntry("PG_DUMP_9_4_PATH is not set", TestEntry{
				arguments:       "--backup --artifact-file /foo --config %s",
				configGenerator: validPgConfig,
				expectedOutput:  "PG_DUMP_9_4_PATH must be set",
			}),
			Entry("PG_RESTORE_9_4_PATH is not set", TestEntry{
				arguments:       "--restore --artifact-file /foo --config %s",
				configGenerator: validPgConfig,
				expectedOutput:  "PG_RESTORE_9_4_PATH must be set",
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
		Context("mysql", func() {
			BeforeEach(func() {
				configFile = buildConfigFile(Config{
					Adapter:  "mysql",
					Username: username,
					Password: password,
					Host:     host,
					Port:     port,
					Database: databaseName,
				})

				fakeDump = binmock.NewBinMock(Fail)
				fakeDump.WhenCalledWith("-V").WillPrintToStdOut("mysqldump  Ver 10.16 Distrib 10.1.24-MariaDB, for Linux (x86_64)")
				fakeDump.WhenCalled().WillExitWith(0)

				fakeClient = binmock.NewBinMock(Fail)
				fakeClient.WhenCalledWith("--skip-column-names",
					"--silent",
					fmt.Sprintf("--user=%s", username),
					fmt.Sprintf("--password=%s", password),
					fmt.Sprintf("--host=%s", host),
					fmt.Sprintf("--port=%d", port),
					`--execute=SELECT VERSION()`).WillPrintToStdOut("10.1.24-MariaDB-wsrep")
			})

			JustBeforeEach(func() {
				cmd := exec.Command(path, "--artifact-file", artifactFile, "--config", configFile.Name(), "--backup")
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

			Context("when 'tables' are specified in the configFile", func() {
				BeforeEach(func() {
					configFile = buildConfigFile(Config{
						Adapter:  "mysql",
						Username: username,
						Password: password,
						Host:     host,
						Port:     port,
						Database: databaseName,
						Tables:   []string{"table1", "table2", "table3"},
					})
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
							"table1",
							"table2",
							"table3",
						}

						Expect(fakeDump.Invocations()[1].Args()).Should(ConsistOf(expectedArgs))
						Expect(fakeDump.Invocations()[1].Env()).Should(HaveKeyWithValue("MYSQL_PWD", password))
					})
				})
			})

			Context("when mysqldump fails", func() {
				BeforeEach(func() {
					fakeDump = binmock.NewBinMock(Fail)
					fakeDump.WhenCalled().WillExitWith(1)
				})

				It("also fails", func() {
					Eventually(session).Should(gexec.Exit(1))
				})
			})

			Context("when mysqldump has a different major version than the server", func() {
				BeforeEach(func() {
					fakeClient = binmock.NewBinMock(Fail)
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
					Expect(string(session.Err.Contents())).Should(ContainSubstring(
						"Version mismatch between mysqldump 10.1.24-MariaDB and the MYSQL server 9.1.24-MariaDB-wsrep"))
				})
			})

			Context("when mysqldump has a different minor version than the server", func() {
				BeforeEach(func() {
					fakeClient = binmock.NewBinMock(Fail)
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
					Expect(string(session.Err.Contents())).Should(ContainSubstring(
						"Version mismatch between mysqldump 10.1.24-MariaDB and the MYSQL server 10.0.24-MariaDB-wsrep"))
				})
			})

			Context("when mysqldump has a different patch version than the server", func() {
				BeforeEach(func() {
					fakeClient = binmock.NewBinMock(Fail)
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
			JustBeforeEach(func() {
				cmd := exec.Command(path, "--artifact-file", artifactFile, "--config", configFile.Name(), "--backup")
				cmd.Env = append(cmd.Env, fmt.Sprintf("PG_DUMP_9_4_PATH=%s", fakeDump.Path))
				cmd.Env = append(cmd.Env, fmt.Sprintf("PG_DUMP_9_6_PATH=%s", fakeDumpPostgres96.Path))
				cmd.Env = append(cmd.Env, fmt.Sprintf("PG_CLIENT_PATH=%s", fakeClient.Path))

				session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
				Eventually(session).Should(gexec.Exit())
			})

			BeforeEach(func() {
				configFile = buildConfigFile(Config{
					Adapter:  "postgres",
					Username: username,
					Password: password,
					Host:     host,
					Port:     port,
					Database: databaseName,
				})

				fakeDump = binmock.NewBinMock(Fail)
				fakeDump.WhenCalled().WillExitWith(0)
				fakeDumpPostgres96 = binmock.NewBinMock(Fail)
				fakeDumpPostgres96.WhenCalled().WillExitWith(0)

				fakeClient = binmock.NewBinMock(Fail)
			})

			Context("9.4", func() {
				BeforeEach(func() {
					fakeClient.WhenCalled().WillPrintToStdOut(" PostgreSQL 9.4.9 on x86_64-unknown-linux-gnu, compiled by gcc (Ubuntu 4.8.4-2ubuntu1~14.04.3) 4.8.4, 64-bit").WillExitWith(0)
				})

				It("calls psql with the correct arguments", func() {
					expectedArgs := []string{
						"--tuples-only",
						fmt.Sprintf("--username=%s", username),
						fmt.Sprintf("--host=%s", host),
						fmt.Sprintf("--port=%d", port),
						databaseName,
						`--command=SELECT VERSION()`,
					}

					Expect(fakeClient.Invocations()).To(HaveLen(1))
					Expect(fakeClient.Invocations()[0].Args()).Should(ConsistOf(expectedArgs))
					Expect(fakeClient.Invocations()[0].Env()).Should(HaveKeyWithValue("PGPASSWORD", password))
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
						fakeDump = binmock.NewBinMock(Fail)
						fakeDump.WhenCalled().WillExitWith(1)
					})

					It("also fails", func() {
						Eventually(session).Should(gexec.Exit(1))
					})
				})
			})

			Context("9.6", func() {
				BeforeEach(func() {
					fakeClient.WhenCalled().WillPrintToStdOut(" PostgreSQL 9.6.3 on x86_64-pc-linux-gnu, compiled by gcc (Ubuntu 4.8.4-2ubuntu1~14.04.3) 4.8.4, 64-bit").WillExitWith(0)
				})

				It("calls psql with the correct arguments", func() {
					expectedArgs := []string{
						"--tuples-only",
						fmt.Sprintf("--username=%s", username),
						fmt.Sprintf("--host=%s", host),
						fmt.Sprintf("--port=%d", port),
						databaseName,
						`--command=SELECT VERSION()`,
					}

					Expect(fakeClient.Invocations()).To(HaveLen(1))
					Expect(fakeClient.Invocations()[0].Args()).Should(ConsistOf(expectedArgs))
					Expect(fakeClient.Invocations()[0].Env()).Should(HaveKeyWithValue("PGPASSWORD", password))
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

					Expect(fakeDumpPostgres96.Invocations()).To(HaveLen(1))
					Expect(fakeDumpPostgres96.Invocations()[0].Args()).Should(ConsistOf(expectedArgs))
					Expect(fakeDumpPostgres96.Invocations()[0].Env()).Should(HaveKeyWithValue("PGPASSWORD", password))
				})

				It("calls pg_dump with the correct env vars", func() {
					Expect(fakeDumpPostgres96.Invocations()).To(HaveLen(1))
					Expect(fakeDumpPostgres96.Invocations()[0].Env()).Should(HaveKeyWithValue("PGPASSWORD", password))
				})

				It("succeeds", func() {
					Expect(session).Should(gexec.Exit(0))
				})

				Context("and pg_dump fails", func() {
					BeforeEach(func() {
						fakeDumpPostgres96 = binmock.NewBinMock(Fail)
						fakeDumpPostgres96.WhenCalled().WillExitWith(1)
					})

					It("also fails", func() {
						Eventually(session).Should(gexec.Exit(1))
					})
				})
			})
		})
	})

	Context("--restore", func() {
		Context("mysql", func() {
			BeforeEach(func() {
				configFile = buildConfigFile(Config{
					Adapter:  "mysql",
					Username: username,
					Password: password,
					Host:     host,
					Port:     port,
					Database: databaseName,
				})

				fakeRestore = binmock.NewBinMock(Fail)
				fakeRestore.WhenCalled().WillExitWith(0)
			})

			JustBeforeEach(func() {
				err := ioutil.WriteFile(artifactFile, []byte("SOME BACKUP SQL"), 0644)
				if err != nil {
					log.Fatalln("Failed to write to artifact file, %s", err)
				}

				cmd := exec.Command(path, "--artifact-file", artifactFile, "--config", configFile.Name(), "--restore")
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
					fakeRestore = binmock.NewBinMock(Fail)
					fakeRestore.WhenCalled().WillExitWith(1)
				})

				It("also fails", func() {
					Eventually(session).Should(gexec.Exit(1))
				})
			})
		})

		Context("postgres", func() {
			BeforeEach(func() {
				configFile = buildConfigFile(Config{
					Adapter:  "postgres",
					Username: username,
					Password: password,
					Host:     host,
					Port:     port,
					Database: databaseName,
				})

				fakeRestore = binmock.NewBinMock(Fail)
				fakeRestore.WhenCalled().WillExitWith(0)
			})

			JustBeforeEach(func() {
				cmd := exec.Command(path, "--artifact-file", artifactFile, "--config", configFile.Name(), "--restore")
				cmd.Env = append(cmd.Env, fmt.Sprintf("PG_RESTORE_9_4_PATH=%s", fakeRestore.Path))

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
					fakeRestore = binmock.NewBinMock(Fail)
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

type Config struct {
	Username string   `json:"username"`
	Password string   `json:"password"`
	Host     string   `json:"host"`
	Port     int      `json:"port"`
	Database string   `json:"database"`
	Adapter  string   `json:"adapter"`
	Tables   []string `json:"tables,omitempty"`
}

func buildConfigFile(config Config) *os.File {
	configFile, err := ioutil.TempFile(os.TempDir(), time.Now().String())
	Expect(err).NotTo(HaveOccurred())

	encoder := json.NewEncoder(configFile)
	err = encoder.Encode(config)
	Expect(err).NotTo(HaveOccurred())

	return configFile
}
