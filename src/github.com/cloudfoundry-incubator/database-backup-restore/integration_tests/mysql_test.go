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

	"strings"

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

		envVars["MYSQL_CLIENT_5_7_PATH"] = fakeMysqlClient.Path
		envVars["MYSQL_DUMP_5_7_PATH"] = fakeMysqlDump.Path
	})

	Context("backup", func() {
		BeforeEach(func() {
			configFile = saveFile(fmt.Sprintf(`{
					"adapter":  "mysql",
					"username": "%s",
					"password": "%s",
					"host":     "%s",
					"port":     %d,
					"database": "%s"
				}`,
				username,
				password,
				host,
				port,
				databaseName))
		})

		JustBeforeEach(func() {
			cmd := exec.Command(
				compiledSDKPath,
				"--artifact-file",
				artifactFile,
				"--config",
				configFile.Name(),
				"--backup")

			for key, val := range envVars {
				cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, val))
			}

			session, err = gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(session).Should(gexec.Exit())
		})

		Context("when mysqldump succeeds", func() {
			BeforeEach(func() {
				fakeMysqlClient.WhenCalled().WillPrintToStdOut("MYSQL server version 5.7.20")
				fakeMysqlDump.WhenCalled().WillExitWith(0)
			})

			It("calls mysql and mysqldump with the correct arguments", func() {
				By("calling mysql to detect the version", func() {
					Expect(fakeMysqlClient.Invocations()).To(HaveLen(1))
					Expect(fakeMysqlClient.Invocations()[0].Args()).Should(ConsistOf(
						fmt.Sprintf("--user=%s", username),
						fmt.Sprintf("--host=%s", host),
						fmt.Sprintf("--port=%d", port),
						"--skip-column-names",
						"--silent",
						`--execute=SELECT VERSION()`,
					))
					Expect(fakeMysqlClient.Invocations()[0].Env()).Should(HaveKeyWithValue("MYSQL_PWD", password))
				})

				By("then calling dump", func() {
					Expect(fakeMysqlDump.Invocations()).To(HaveLen(1))
					expectedArgs := []string{
						fmt.Sprintf("--user=%s", username),
						fmt.Sprintf("--host=%s", host),
						fmt.Sprintf("--port=%d", port),
						"-v",
						"--single-transaction",
						"--skip-add-locks",
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
					configFile = saveFile(fmt.Sprintf(`{
					"adapter":  "mysql",
					"username": "%s",
					"password": "%s",
					"host":     "%s",
					"port":     %d,
					"database": "%s",
					"tables": ["table1", "table2", "table3"]
				}`,
						username,
						password,
						host,
						port,
						databaseName))
				})

				It("calls mysqldump with the correct arguments", func() {
					expectedArgs := []string{
						fmt.Sprintf("--user=%s", username),
						fmt.Sprintf("--host=%s", host),
						fmt.Sprintf("--port=%d", port),
						"-v",
						"--single-transaction",
						"--skip-add-locks",
						fmt.Sprintf("--result-file=%s", artifactFile),
						databaseName,
						"table1",
						"table2",
						"table3",
					}

					Expect(fakeMysqlDump.Invocations()[0].Args()).Should(ConsistOf(expectedArgs))
				})
			})

			Context("when TLS is configured", func() {
				BeforeEach(func() {
					configFile = saveFile(fmt.Sprintf(`{
							"adapter":  "mysql",
							"username": "%s",
							"password": "%s",
							"host":     "%s",
							"port":     %d,
							"database": "%s",
							"tls": {
								"cert": {
									"ca": "A_CA_CERT"
								}
							}
						}`,
						username,
						password,
						host,
						port,
						databaseName))
				})

				It("calls mysql and mysqldump with the correct arguments", func() {
					By("calling mysql to detect the version", func() {
						Expect(fakeMysqlClient.Invocations()).To(HaveLen(1))
						Expect(fakeMysqlClient.Invocations()[0].Args()).Should(ConsistOf(
							fmt.Sprintf("--user=%s", username),
							fmt.Sprintf("--host=%s", host),
							fmt.Sprintf("--port=%d", port),
							HavePrefix("--ssl-ca="),
							"--ssl-mode=VERIFY_IDENTITY",
							"--skip-column-names",
							"--silent",
							`--execute=SELECT VERSION()`,
						))
						Expect(fakeMysqlClient.Invocations()[0].Env()).Should(HaveKeyWithValue("MYSQL_PWD", password))
					})

					By("then calling dump", func() {
						expectedArgs := []interface{}{
							fmt.Sprintf("--user=%s", username),
							fmt.Sprintf("--host=%s", host),
							fmt.Sprintf("--port=%d", port),
							HavePrefix("--ssl-ca="),
							"--ssl-mode=VERIFY_IDENTITY",
							"-v",
							"--single-transaction",
							"--skip-add-locks",
							fmt.Sprintf("--result-file=%s", artifactFile),
							databaseName,
						}

						Expect(fakeMysqlDump.Invocations()[0].Args()).Should(ConsistOf(expectedArgs))

						caCertPath := strings.Split(fakeMysqlDump.Invocations()[0].Args()[3], "=")[1]
						Expect(caCertPath).To(BeAnExistingFile())
						Expect(ioutil.ReadFile(caCertPath)).To(Equal([]byte("A_CA_CERT")))
					})
				})
			})
		})

		Context("when version detection fails", func() {
			BeforeEach(func() {
				fakeMysqlClient.WhenCalledWith(
					fmt.Sprintf("--user=%s", username),
					fmt.Sprintf("--host=%s", host),
					fmt.Sprintf("--port=%d", port),
					"--skip-column-names",
					"--silent",
					`--execute=SELECT VERSION()`,
				).WillExitWith(1).WillPrintToStdErr("VERSION DETECTION FAILED!")
			})

			It("also fails", func() {
				Expect(fakeMysqlClient.Invocations()[0].Env()).Should(HaveKeyWithValue("MYSQL_PWD", password))
				Eventually(session).Should(gexec.Exit(1))
				Expect(session.Err).To(gbytes.Say("VERSION DETECTION FAILED!"))
			})
		})

		Context("when mysqldump fails", func() {
			BeforeEach(func() {
				fakeMysqlClient.WhenCalledWith(
					fmt.Sprintf("--user=%s", username),
					fmt.Sprintf("--host=%s", host),
					fmt.Sprintf("--port=%d", port),
					"--skip-column-names",
					"--silent",
					`--execute=SELECT VERSION()`,
				).WillPrintToStdOut("MYSQL server version 5.7.20")
				fakeMysqlDump.WhenCalled().WillExitWith(1)
			})

			It("also fails", func() {
				Expect(fakeMysqlClient.Invocations()[0].Env()).Should(HaveKeyWithValue("MYSQL_PWD", password))
				Eventually(session).Should(gexec.Exit(1))
			})
		})

		Context("when the server has an unsupported mysql major version", func() {
			BeforeEach(func() {
				fakeMysqlClient.WhenCalledWith(
					fmt.Sprintf("--user=%s", username),
					fmt.Sprintf("--host=%s", host),
					fmt.Sprintf("--port=%d", port),
					"--skip-column-names",
					"--silent",
					`--execute=SELECT VERSION()`,
				).WillPrintToStdOut("MYSQL server version 4.7.20")
			})

			It("fails because of a version mismatch", func() {
				Expect(fakeMysqlClient.Invocations()[0].Env()).Should(HaveKeyWithValue("MYSQL_PWD", password))
				Expect(session).Should(gexec.Exit(1))
				Expect(string(session.Err.Contents())).Should(ContainSubstring(
					"unsupported version of mysql: 4.7"),
				)
			})
		})

		Context("when the server has an unsupported mysql minor version", func() {
			BeforeEach(func() {
				fakeMysqlClient.WhenCalledWith(
					fmt.Sprintf("--user=%s", username),
					fmt.Sprintf("--host=%s", host),
					fmt.Sprintf("--port=%d", port),
					"--skip-column-names",
					"--silent",
					`--execute=SELECT VERSION()`,
				).WillPrintToStdOut("MYSQL server version 5.9.20")
			})

			It("fails because of a version mismatch", func() {
				Expect(fakeMysqlClient.Invocations()[0].Env()).Should(HaveKeyWithValue("MYSQL_PWD", password))
				Expect(session).Should(gexec.Exit(1))
				Expect(string(session.Err.Contents())).Should(ContainSubstring(
					"unsupported version of mysql: 5.9"),
				)
			})
		})

		Context("when the server has a supported mysql minor version with a different patch to the packaged utility", func() {
			BeforeEach(func() {
				fakeMysqlDump.WhenCalled().WillExitWith(0)
				fakeMysqlClient.WhenCalledWith(
					fmt.Sprintf("--user=%s", username),
					fmt.Sprintf("--host=%s", host),
					fmt.Sprintf("--port=%d", port),
					"--skip-column-names",
					"--silent",
					`--execute=SELECT VERSION()`,
				).WillPrintToStdOut("MYSQL server version 5.7.5")
			})

			It("succeeds despite different patch versions", func() {
				Expect(fakeMysqlClient.Invocations()[0].Env()).Should(HaveKeyWithValue("MYSQL_PWD", password))
				Expect(session).Should(gexec.Exit(0))
			})
		})
	})

	Context("restore", func() {
		BeforeEach(func() {
			configFile = saveFile(fmt.Sprintf(`{
					"adapter":  "mysql",
					"username": "%s",
					"password": "%s",
					"host":     "%s",
					"port":     %d,
					"database": "%s"
				}`,
				username,
				password,
				host,
				port,
				databaseName))
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

		Context("when mysql succeeds", func() {
			BeforeEach(func() {
				fakeMysqlClient.WhenCalledWith(
					fmt.Sprintf("--user=%s", username),
					fmt.Sprintf("--host=%s", host),
					fmt.Sprintf("--port=%d", port),
					"--skip-column-names",
					"--silent",
					`--execute=SELECT VERSION()`,
				).WillPrintToStdOut("MYSQL server version 5.7.20")
				fakeMysqlClient.WhenCalled().WillExitWith(0)
			})

			It("calls mysql with the correct arguments", func() {
				expectedArgs := []string{
					"-v",
					fmt.Sprintf("--user=%s", username),
					fmt.Sprintf("--host=%s", host),
					fmt.Sprintf("--port=%d", port),
					databaseName,
				}

				Expect(fakeMysqlClient.Invocations()).To(HaveLen(2))
				Expect(fakeMysqlClient.Invocations()[0].Env()).Should(HaveKeyWithValue("MYSQL_PWD", password))
				Expect(fakeMysqlClient.Invocations()[1].Args()).Should(ConsistOf(expectedArgs))
				Expect(fakeMysqlClient.Invocations()[1].Stdin()).Should(ConsistOf("SOME BACKUP SQL"))
				Expect(fakeMysqlClient.Invocations()[1].Env()).Should(HaveKeyWithValue("MYSQL_PWD", password))
			})

			It("succeeds", func() {
				Expect(session).Should(gexec.Exit(0))
			})
		})

		Context("when mysql fails", func() {
			BeforeEach(func() {
				fakeMysqlClient.WhenCalledWith(
					fmt.Sprintf("--user=%s", username),
					fmt.Sprintf("--host=%s", host),
					fmt.Sprintf("--port=%d", port),
					"--skip-column-names",
					"--silent",
					`--execute=SELECT VERSION()`,
				).WillPrintToStdOut("MYSQL server version 5.7.20")
				fakeMysqlClient.WhenCalled().WillExitWith(1)
			})

			It("also fails", func() {
				Expect(fakeMysqlClient.Invocations()[0].Env()).Should(HaveKeyWithValue("MYSQL_PWD", password))
				Eventually(session).Should(gexec.Exit(1))
			})
		})
	})
})
