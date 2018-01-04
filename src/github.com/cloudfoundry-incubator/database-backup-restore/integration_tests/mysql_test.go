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
)

var _ = Describe("MySQL", func() {
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
		fakeMysqlDump.Reset()
		fakeMysqlClient.Reset()
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

			envVars["MARIADB_DUMP_PATH"] = fakeMysqlDump.Path

		})

		JustBeforeEach(func() {
			cmd := exec.Command(
				compiledSDKPath,
				"--artifact-file",
				artifactFile,
				"--config",
				configFile.Name(),
				"--backup")
			envVars["MARIADB_CLIENT_PATH"] = fakeMysqlClient.Path

			for key, val := range envVars {
				cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, val))
			}

			session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(session).Should(gexec.Exit())
		})

		Context("MARIADB_DUMP_PATH env var is missing", func() {
			BeforeEach(func() {
				delete(envVars, "MARIADB_DUMP_PATH")
			})
			It("raises an appropriate error", func() {
				Expect(session.Err).To(gbytes.Say("MARIADB_DUMP_PATH must be set"))
			})
		})

		Context("MARIADB_DUMP_PATH is set", func() {
			Context("when mysqldump succeeds", func() {
				BeforeEach(func() {
					fakeMysqlClient.WhenCalledWith(
						"--skip-column-names",
						"--silent",
						fmt.Sprintf("--user=%s", username),
						fmt.Sprintf("--password=%s", password),
						fmt.Sprintf("--host=%s", host),
						fmt.Sprintf("--port=%d", port),
						`--execute=SELECT VERSION()`,
					).WillPrintToStdOut("10.1.24-MariaDB-wsrep")
					fakeMysqlClient.WhenCalledWith(
						"--skip-column-names",
						"--silent",
						fmt.Sprintf("--user=%s", username),
						fmt.Sprintf("--password=%s", password),
						fmt.Sprintf("--host=%s", host),
						fmt.Sprintf("--port=%d", port),
						`--execute=SELECT @@VERSION_COMMENT`,
					).WillPrintToStdOut("MariaDB Server")
					fakeMysqlDump.WhenCalled().WillExitWith(0)
				})
				It("calls mysqldump with the correct arguments", func() {
					Expect(fakeMysqlDump.Invocations()).To(HaveLen(1))

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

						Expect(fakeMysqlDump.Invocations()[0].Args()).Should(ConsistOf(expectedArgs))
						Expect(fakeMysqlDump.Invocations()[0].Env()).Should(HaveKeyWithValue("MYSQL_PWD", password))
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

							Expect(fakeMysqlDump.Invocations()[0].Args()).Should(ConsistOf(expectedArgs))
						})
					})
				})
			})

			Context("when mysqldump fails", func() {
				BeforeEach(func() {
					fakeMysqlClient.WhenCalledWith(
						"--skip-column-names",
						"--silent",
						fmt.Sprintf("--user=%s", username),
						fmt.Sprintf("--password=%s", password),
						fmt.Sprintf("--host=%s", host),
						fmt.Sprintf("--port=%d", port),
						`--execute=SELECT VERSION()`,
					).WillPrintToStdOut("10.1.24-MariaDB-wsrep")
					fakeMysqlClient.WhenCalledWith(
						"--skip-column-names",
						"--silent",
						fmt.Sprintf("--user=%s", username),
						fmt.Sprintf("--password=%s", password),
						fmt.Sprintf("--host=%s", host),
						fmt.Sprintf("--port=%d", port),
						`--execute=SELECT @@VERSION_COMMENT`,
					).WillPrintToStdOut("MariaDB Server")
					fakeMysqlDump.WhenCalled().WillExitWith(1)
				})

				It("also fails", func() {
					Eventually(session).Should(gexec.Exit(1))
				})
			})

			Context("when the server has an unsupported mariadb major version", func() {
				BeforeEach(func() {
					fakeMysqlClient.WhenCalledWith(
						"--skip-column-names",
						"--silent",
						fmt.Sprintf("--user=%s", username),
						fmt.Sprintf("--password=%s", password),
						fmt.Sprintf("--host=%s", host),
						fmt.Sprintf("--port=%d", port),
						`--execute=SELECT VERSION()`,
					).WillPrintToStdOut("9.1.24-MariaDB-wsrep")
					fakeMysqlClient.WhenCalledWith(
						"--skip-column-names",
						"--silent",
						fmt.Sprintf("--user=%s", username),
						fmt.Sprintf("--password=%s", password),
						fmt.Sprintf("--host=%s", host),
						fmt.Sprintf("--port=%d", port),
						`--execute=SELECT @@VERSION_COMMENT`,
					).WillPrintToStdOut("MariaDB Server")
				})

				It("fails because of a version mismatch", func() {
					Expect(session).Should(gexec.Exit(1))
					Expect(string(session.Err.Contents())).Should(ContainSubstring(
						"unsupported version of mariadb: 9.1"),
					)
				})
			})

			Context("when the server has an unsupported mariadb minor version", func() {
				BeforeEach(func() {
					fakeMysqlClient.WhenCalledWith(
						"--skip-column-names",
						"--silent",
						fmt.Sprintf("--user=%s", username),
						fmt.Sprintf("--password=%s", password),
						fmt.Sprintf("--host=%s", host),
						fmt.Sprintf("--port=%d", port),
						`--execute=SELECT VERSION()`,
					).WillPrintToStdOut("10.0.24-MariaDB-wsrep")
					fakeMysqlClient.WhenCalledWith(
						"--skip-column-names",
						"--silent",
						fmt.Sprintf("--user=%s", username),
						fmt.Sprintf("--password=%s", password),
						fmt.Sprintf("--host=%s", host),
						fmt.Sprintf("--port=%d", port),
						`--execute=SELECT @@VERSION_COMMENT`,
					).WillPrintToStdOut("MariaDB Server")
				})

				It("fails because of a version mismatch", func() {
					Expect(session).Should(gexec.Exit(1))
					Expect(string(session.Err.Contents())).Should(ContainSubstring(
						"unsupported version of mariadb: 10.0"),
					)
				})
			})

			Context("when the server has a supported mariadb minor version with a different patch to the packaged utility", func() {
				BeforeEach(func() {
					fakeMysqlDump.WhenCalled().WillExitWith(0)
					fakeMysqlClient.WhenCalledWith(
						"--skip-column-names",
						"--silent",
						fmt.Sprintf("--user=%s", username),
						fmt.Sprintf("--password=%s", password),
						fmt.Sprintf("--host=%s", host),
						fmt.Sprintf("--port=%d", port),
						`--execute=SELECT VERSION()`,
					).WillPrintToStdOut("10.1.22-MariaDB-wsrep")
					fakeMysqlClient.WhenCalledWith(
						"--skip-column-names",
						"--silent",
						fmt.Sprintf("--user=%s", username),
						fmt.Sprintf("--password=%s", password),
						fmt.Sprintf("--host=%s", host),
						fmt.Sprintf("--port=%d", port),
						`--execute=SELECT @@VERSION_COMMENT`,
					).WillPrintToStdOut("MariaDB Server")
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

		Context("MARIADB_CLIENT_PATH env var is missing", func() {
			BeforeEach(func() {
				delete(envVars, "MARIADB_CLIENT_PATH")
			})

			It("raises an appropriate error", func() {
				Expect(session.Err).To(gbytes.Say("MARIADB_CLIENT_PATH must be set"))
			})
		})

		Context("MARIADB_CLIENT_PATH is set", func() {
			BeforeEach(func() {
				fakeMysqlClient.WhenCalledWith(
					"--skip-column-names",
					"--silent",
					fmt.Sprintf("--user=%s", username),
					fmt.Sprintf("--password=%s", password),
					fmt.Sprintf("--host=%s", host),
					fmt.Sprintf("--port=%d", port),
					`--execute=SELECT VERSION()`,
				).WillPrintToStdOut("10.1.24-MariaDB-wsrep")
				fakeMysqlClient.WhenCalledWith(
					"--skip-column-names",
					"--silent",
					fmt.Sprintf("--user=%s", username),
					fmt.Sprintf("--password=%s", password),
					fmt.Sprintf("--host=%s", host),
					fmt.Sprintf("--port=%d", port),
					`--execute=SELECT @@VERSION_COMMENT`,
				).WillPrintToStdOut("MariaDB Server")
				fakeMysqlClient.WhenCalled().WillExitWith(0)
				envVars["MARIADB_CLIENT_PATH"] = fakeMysqlClient.Path
			})

			It("calls mysql with the correct arguments", func() {
				expectedArgs := []string{
					"-v",
					fmt.Sprintf("--user=%s", username),
					fmt.Sprintf("--host=%s", host),
					fmt.Sprintf("--port=%d", port),
					databaseName,
				}

				Expect(fakeMysqlClient.Invocations()).To(HaveLen(3))
				Expect(fakeMysqlClient.Invocations()[2].Args()).Should(ConsistOf(expectedArgs))
				Expect(fakeMysqlClient.Invocations()[2].Stdin()).Should(ConsistOf("SOME BACKUP SQL"))
				Expect(fakeMysqlClient.Invocations()[2].Env()).Should(HaveKeyWithValue("MYSQL_PWD", password))
			})

			It("succeeds", func() {
				Expect(session).Should(gexec.Exit(0))
			})

		})
		Context("and mysql fails", func() {
			BeforeEach(func() {
				fakeMysqlClient.WhenCalledWith(
					"--skip-column-names",
					"--silent",
					fmt.Sprintf("--user=%s", username),
					fmt.Sprintf("--password=%s", password),
					fmt.Sprintf("--host=%s", host),
					fmt.Sprintf("--port=%d", port),
					`--execute=SELECT VERSION()`,
				).WillPrintToStdOut("10.1.24-MariaDB-wsrep")
				fakeMysqlClient.WhenCalledWith(
					"--skip-column-names",
					"--silent",
					fmt.Sprintf("--user=%s", username),
					fmt.Sprintf("--password=%s", password),
					fmt.Sprintf("--host=%s", host),
					fmt.Sprintf("--port=%d", port),
					`--execute=SELECT @@VERSION_COMMENT`,
				).WillPrintToStdOut("MariaDB Server")
				fakeMysqlClient.WhenCalled().WillExitWith(1)
				envVars["MARIADB_CLIENT_PATH"] = fakeMysqlClient.Path
			})

			It("also fails", func() {
				Eventually(session).Should(gexec.Exit(1))
			})
		})
	})
})
