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

package integration_tests

import (
	"fmt"
	"os/exec"

	"io/ioutil"
	"log"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	binmock "github.com/pivotal-cf-experimental/go-binmock"
)

var _ = Describe("MySQL", func() {
	var fakeMysqlDump *binmock.Mock
	var mysqlClient *binmock.Mock
	var session *gexec.Session
	var username = "testuser"
	var host = "127.0.0.1"
	var port = 1234
	var databaseName = "mycooldb"
	var password = "password"
	var artifactFile string
	var err error
	var configFile *os.File

	BeforeEach(func() {
		artifactFile = tempFilePath()
	})

	Context("backup", func() {

		BeforeEach(func() {
			configFile = buildConfigFile(Config{
				Adapter:  "mysql",
				Username: username,
				Password: password,
				Host:     host,
				Port:     port,
				Database: databaseName,
			})

			fakeMysqlDump = binmock.NewBinMock(Fail)
			fakeMysqlDump.WhenCalledWith("-V").
				WillPrintToStdOut("mysqldump  Ver 10.16 Distrib 10.1.24-MariaDB, for Linux (x86_64)")
			fakeMysqlDump.WhenCalled().WillExitWith(0)

			mysqlClient = binmock.NewBinMock(Fail)
			mysqlClient.WhenCalledWith(
				"--skip-column-names",
				"--silent",
				fmt.Sprintf("--user=%s", username),
				fmt.Sprintf("--password=%s", password),
				fmt.Sprintf("--host=%s", host),
				fmt.Sprintf("--port=%d", port),
				`--execute=SELECT VERSION()`,
			).WillPrintToStdOut("10.1.24-MariaDB-wsrep")

		})

		JustBeforeEach(func() {
			cmd := exec.Command(
				compiledSDKPath,
				"--artifact-file",
				artifactFile,
				"--config",
				configFile.Name(),
				"--backup")
			envVars["MYSQL_CLIENT_PATH"] = mysqlClient.Path

			for key, val := range envVars {
				cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, val))
			}

			session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(session).Should(gexec.Exit())
		})

		Context("MYSQL_DUMP_PATH env var is missing", func() {
			BeforeEach(func() {
				delete(envVars, "MYSQL_DUMP_PATH")
			})
			It("raises an appropriate error", func() {
				Expect(session.Err).To(gbytes.Say("MYSQL_DUMP_PATH must be set"))
			})
		})

		Context("MYSQL_DUMP_PATH is set", func() {
			BeforeEach(func() {
				envVars["MYSQL_DUMP_PATH"] = fakeMysqlDump.Path
			})
			It("calls mysqldump with the correct arguments", func() {
				Expect(fakeMysqlDump.Invocations()).To(HaveLen(2))

				By("first checking the version", func() {
					Expect(fakeMysqlDump.Invocations()[0].Args()).Should(ConsistOf("-V"))
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

					Expect(fakeMysqlDump.Invocations()[1].Args()).Should(ConsistOf(expectedArgs))
					Expect(fakeMysqlDump.Invocations()[1].Env()).Should(HaveKeyWithValue("MYSQL_PWD", password))
				})

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

						Expect(fakeMysqlDump.Invocations()[1].Args()).Should(ConsistOf(expectedArgs))
					})
				})
			})

			Context("when mysqldump fails", func() {
				BeforeEach(func() {
					fakeMysqlDump = binmock.NewBinMock(Fail)
					fakeMysqlDump.WhenCalled().WillExitWith(1)
					envVars["MYSQL_DUMP_PATH"] = fakeMysqlDump.Path
				})

				It("also fails", func() {
					Eventually(session).Should(gexec.Exit(1))
				})
			})

			Context("when mysqldump has a different major version than the server", func() {
				BeforeEach(func() {
					mysqlClient = binmock.NewBinMock(Fail)
					mysqlClient.WhenCalledWith(
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
						"Version mismatch between dump utility 10.1.24-MariaDB and " +
							"the database server 9.1.24-MariaDB-wsrep"),
					)
				})
			})

			Context("when mysqldump has a different minor version than the server", func() {
				BeforeEach(func() {
					mysqlClient = binmock.NewBinMock(Fail)
					mysqlClient.WhenCalledWith(
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
						"Version mismatch between dump utility 10.1.24-MariaDB and " +
							"the database server 10.0.24-MariaDB-wsrep"),
					)
				})
			})

			Context("when mysqldump has a different patch version than the server", func() {
				BeforeEach(func() {
					mysqlClient = binmock.NewBinMock(Fail)
					mysqlClient.WhenCalledWith(
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

	})

	Context("restore", func() {
		BeforeEach(func() {
			configFile = buildConfigFile(Config{
				Adapter:  "mysql",
				Username: username,
				Password: password,
				Host:     host,
				Port:     port,
				Database: databaseName,
			})

			mysqlClient = binmock.NewBinMock(Fail)
			mysqlClient.WhenCalled().WillExitWith(0)
		})

		JustBeforeEach(func() {
			err := ioutil.WriteFile(artifactFile, []byte("SOME BACKUP SQL"), 0644)
			if err != nil {
				log.Fatalln("Failed to write to artifact file, %s", err)
			}

			cmd := exec.Command(
				compiledSDKPath,
				"--artifact-file",
				artifactFile,
				"--config",
				configFile.Name(),
				"--restore")

			for key, val := range envVars {
				cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, val))
			}

			session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(session).Should(gexec.Exit())
		})

		Context("MYSQL_CLIENT_PATH env var is missing", func() {
			BeforeEach(func() {
				delete(envVars, "MYSQL_CLIENT_PATH")
			})

			It("raises an appropriate error", func() {
				Expect(session.Err).To(gbytes.Say("MYSQL_CLIENT_PATH must be set"))
			})
		})

		Context("MYSQL_CLIENT_PATH is set", func() {
			BeforeEach(func() {
				envVars["MYSQL_CLIENT_PATH"] = mysqlClient.Path
			})

			It("calls mysql with the correct arguments", func() {
				expectedArgs := []string{
					"-v",
					fmt.Sprintf("--user=%s", username),
					fmt.Sprintf("--host=%s", host),
					fmt.Sprintf("--port=%d", port),
					databaseName,
				}

				Expect(mysqlClient.Invocations()).To(HaveLen(1))
				Expect(mysqlClient.Invocations()[0].Args()).Should(ConsistOf(expectedArgs))
				Expect(mysqlClient.Invocations()[0].Stdin()).Should(ConsistOf("SOME BACKUP SQL"))
				Expect(mysqlClient.Invocations()[0].Env()).Should(HaveKeyWithValue("MYSQL_PWD", password))
			})

			It("succeeds", func() {
				Expect(session).Should(gexec.Exit(0))
			})

			Context("and mysql fails", func() {
				BeforeEach(func() {
					mysqlClient = binmock.NewBinMock(Fail)
					mysqlClient.WhenCalled().WillExitWith(1)
					envVars["MYSQL_CLIENT_PATH"] = mysqlClient.Path
				})

				It("also fails", func() {
					Eventually(session).Should(gexec.Exit(1))
				})
			})
		})
	})
})
