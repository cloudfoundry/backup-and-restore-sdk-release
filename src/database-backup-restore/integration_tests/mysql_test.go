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
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"

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

	Context("mysql 5.7", func() {
		BeforeEach(func() {
			artifactFile = tempFilePath()
			fakeMysqlDump57.Reset()
			fakeMysqlClient57.Reset()

			envVars["MYSQL_CLIENT_5_7_PATH"] = fakeMysqlClient57.Path
			envVars["MYSQL_DUMP_5_7_PATH"] = fakeMysqlDump57.Path
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
					fakeMysqlClient57.WhenCalled().WillPrintToStdOut("MYSQL server version 5.7.27")
					fakeMysqlDump57.WhenCalled().WillExitWith(0)
				})

				It("calls mysql and mysqldump with the correct arguments", func() {
					By("calling mysql to detect the version", func() {
						Expect(fakeMysqlClient57.Invocations()).To(HaveLen(1))
						Expect(fakeMysqlClient57.Invocations()[0].Args()).Should(ConsistOf(
							fmt.Sprintf("--user=%s", username),
							fmt.Sprintf("--host=%s", host),
							fmt.Sprintf("--port=%d", port),
							"--skip-column-names",
							"--silent",
							`--execute=SELECT VERSION()`,
						))
						Expect(fakeMysqlClient57.Invocations()[0].Env()).Should(HaveKeyWithValue("MYSQL_PWD", password))
					})

					By("then calling dump", func() {
						Expect(fakeMysqlDump57.Invocations()).To(HaveLen(1))
						expectedArgs := []interface{}{
							fmt.Sprintf("--user=%s", username),
							fmt.Sprintf("--host=%s", host),
							fmt.Sprintf("--port=%d", port),
							"-v",
							"--set-gtid-purged=OFF",
							"--single-transaction",
							"--skip-add-locks",
							fmt.Sprintf("--result-file=%s", artifactFile),
							databaseName,
						}

						Expect(fakeMysqlDump57.Invocations()[0].Args()).Should(ConsistOf(expectedArgs))
						Expect(fakeMysqlDump57.Invocations()[0].Env()).Should(HaveKeyWithValue("MYSQL_PWD", password))
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
						expectedArgs := []interface{}{
							fmt.Sprintf("--user=%s", username),
							fmt.Sprintf("--host=%s", host),
							fmt.Sprintf("--port=%d", port),
							"-v",
							"--single-transaction",
							"--set-gtid-purged=OFF",
							"--skip-add-locks",
							fmt.Sprintf("--result-file=%s", artifactFile),
							databaseName,
							"table1",
							"table2",
							"table3",
						}

						Expect(fakeMysqlDump57.Invocations()[0].Args()).Should(ConsistOf(expectedArgs))
					})
				})

				Context("when TLS is configured with hostname verification turned off", func() {
					BeforeEach(func() {
						configFile = saveFile(fmt.Sprintf(`{
							"adapter":  "mysql",
							"username": "%s",
							"password": "%s",
							"host":     "%s",
							"port":     %d,
							"database": "%s",
							"tls": {
								"skip_host_verify": true,
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
							Expect(fakeMysqlClient57.Invocations()).To(HaveLen(1))
							Expect(fakeMysqlClient57.Invocations()[0].Args()).Should(ConsistOf(
								fmt.Sprintf("--user=%s", username),
								fmt.Sprintf("--host=%s", host),
								fmt.Sprintf("--port=%d", port),
								HavePrefix("--ssl-ca="),
								"--ssl-mode=VERIFY_CA",
								"--skip-column-names",
								"--silent",
								`--execute=SELECT VERSION()`,
							))
							Expect(fakeMysqlClient57.Invocations()[0].Env()).Should(
								HaveKeyWithValue("MYSQL_PWD", password))
						})

						By("then calling dump", func() {
							expectedArgs := []interface{}{
								fmt.Sprintf("--user=%s", username),
								fmt.Sprintf("--host=%s", host),
								fmt.Sprintf("--port=%d", port),
								HavePrefix("--ssl-ca="),
								"--ssl-mode=VERIFY_CA",
								"--set-gtid-purged=OFF",
								"-v",
								"--single-transaction",
								"--skip-add-locks",
								fmt.Sprintf("--result-file=%s", artifactFile),
								databaseName,
							}

							Expect(fakeMysqlDump57.Invocations()[0].Args()).Should(ConsistOf(expectedArgs))
						})
					})
				})

				Context("when TLS is configured with client cert and private key", func() {
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
									"ca": "A_CA_CERT",
									"certificate": "A_CLIENT_CERT",
									"private_key": "A_CLIENT_KEY"
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
							Expect(fakeMysqlClient57.Invocations()).To(HaveLen(1))
							Expect(fakeMysqlClient57.Invocations()[0].Args()).Should(ConsistOf(
								fmt.Sprintf("--user=%s", username),
								fmt.Sprintf("--host=%s", host),
								fmt.Sprintf("--port=%d", port),
								HavePrefix("--ssl-ca="),
								HavePrefix("--ssl-cert="),
								HavePrefix("--ssl-key="),
								"--ssl-mode=VERIFY_IDENTITY",
								"--skip-column-names",
								"--silent",
								`--execute=SELECT VERSION()`,
							))
							Expect(fakeMysqlClient57.Invocations()[0].Env()).Should(
								HaveKeyWithValue("MYSQL_PWD", password))
						})

						By("then calling dump", func() {
							expectedArgs := []interface{}{
								fmt.Sprintf("--user=%s", username),
								fmt.Sprintf("--host=%s", host),
								fmt.Sprintf("--port=%d", port),
								HavePrefix("--ssl-ca="),
								HavePrefix("--ssl-cert="),
								HavePrefix("--ssl-key="),
								"--ssl-mode=VERIFY_IDENTITY",
								"-v",
								"--single-transaction",
								"--skip-add-locks",
								"--set-gtid-purged=OFF",
								fmt.Sprintf("--result-file=%s", artifactFile),
								databaseName,
							}

							Expect(fakeMysqlDump57.Invocations()[0].Args()).Should(ConsistOf(expectedArgs))
						})
					})
				})
			})

			Context("when version detection fails", func() {
				BeforeEach(func() {
					fakeMysqlClient57.WhenCalledWith(
						fmt.Sprintf("--user=%s", username),
						fmt.Sprintf("--host=%s", host),
						fmt.Sprintf("--port=%d", port),
						"--skip-column-names",
						"--silent",
						`--execute=SELECT VERSION()`,
					).WillExitWith(1).WillPrintToStdErr("VERSION DETECTION FAILED!")
				})

				It("also fails", func() {
					Expect(fakeMysqlClient57.Invocations()[0].Env()).Should(HaveKeyWithValue("MYSQL_PWD", password))
					Eventually(session).Should(gexec.Exit(1))
					Expect(session.Err).To(gbytes.Say("VERSION DETECTION FAILED!"))
				})
			})

			Context("when mysqldump fails", func() {
				BeforeEach(func() {
					fakeMysqlClient57.WhenCalledWith(
						fmt.Sprintf("--user=%s", username),
						fmt.Sprintf("--host=%s", host),
						fmt.Sprintf("--port=%d", port),
						"--skip-column-names",
						"--silent",
						`--execute=SELECT VERSION()`,
					).WillPrintToStdOut("MYSQL server version 5.6.38")
					fakeMysqlDump57.WhenCalled().WillExitWith(1)
				})

				It("also fails", func() {
					Expect(fakeMysqlClient57.Invocations()[0].Env()).Should(HaveKeyWithValue("MYSQL_PWD", password))
					Eventually(session).Should(gexec.Exit(1))
				})
			})

			Context("when the server has an unsupported mysql major version", func() {
				BeforeEach(func() {
					fakeMysqlClient57.WhenCalledWith(
						fmt.Sprintf("--user=%s", username),
						fmt.Sprintf("--host=%s", host),
						fmt.Sprintf("--port=%d", port),
						"--skip-column-names",
						"--silent",
						`--execute=SELECT VERSION()`,
					).WillPrintToStdOut("MYSQL server version 4.7.20")
				})

				It("fails because of a version mismatch", func() {
					Expect(fakeMysqlClient57.Invocations()[0].Env()).Should(HaveKeyWithValue("MYSQL_PWD", password))
					Expect(session).Should(gexec.Exit(1))
					Expect(string(session.Err.Contents())).Should(ContainSubstring(
						"unsupported version of mysql: 4.7"),
					)
				})
			})
			Context("when the server has a previously supported mysql version", func() {
				BeforeEach(func() {
					fakeMysqlClient57.WhenCalledWith(
						fmt.Sprintf("--user=%s", username),
						fmt.Sprintf("--host=%s", host),
						fmt.Sprintf("--port=%d", port),
						"--skip-column-names",
						"--silent",
						`--execute=SELECT VERSION()`,
					).WillPrintToStdOut("MYSQL server version 5.6.1")
				})

				It("fails because of a version mismatch", func() {
					Expect(fakeMysqlClient57.Invocations()[0].Env()).Should(HaveKeyWithValue("MYSQL_PWD", password))
					Expect(session).Should(gexec.Exit(1))
					Expect(string(session.Err.Contents())).Should(ContainSubstring(
						"unsupported version of mysql: 5.6"),
					)
				})
			})

			Context("when the server has an unsupported mysql minor version", func() {
				BeforeEach(func() {
					fakeMysqlClient57.WhenCalledWith(
						fmt.Sprintf("--user=%s", username),
						fmt.Sprintf("--host=%s", host),
						fmt.Sprintf("--port=%d", port),
						"--skip-column-names",
						"--silent",
						`--execute=SELECT VERSION()`,
					).WillPrintToStdOut("MYSQL server version 5.9.20")
				})

				It("fails because of a version mismatch", func() {
					Expect(fakeMysqlClient57.Invocations()[0].Env()).Should(HaveKeyWithValue("MYSQL_PWD", password))
					Expect(session).Should(gexec.Exit(1))
					Expect(string(session.Err.Contents())).Should(ContainSubstring(
						"unsupported version of mysql: 5.9"),
					)
				})
			})

			Context("when the server has a supported mysql minor version with a different patch to the packaged utility", func() {
				BeforeEach(func() {
					fakeMysqlDump57.WhenCalled().WillExitWith(0)
					fakeMysqlClient57.WhenCalledWith(
						fmt.Sprintf("--user=%s", username),
						fmt.Sprintf("--host=%s", host),
						fmt.Sprintf("--port=%d", port),
						"--skip-column-names",
						"--silent",
						`--execute=SELECT VERSION()`,
					).WillPrintToStdOut("MYSQL server version 5.7.27")
				})

				It("succeeds despite different patch versions", func() {
					Expect(fakeMysqlClient57.Invocations()[0].Env()).Should(HaveKeyWithValue("MYSQL_PWD", password))
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
				err := os.WriteFile(artifactFile, []byte("SOME BACKUP SQL"), 0644)
				Expect(err).ToNot(HaveOccurred())

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
					fakeMysqlClient57.WhenCalled().WillPrintToStdOut("MYSQL server version 5.7.27")
					fakeMysqlClient57.WhenCalled().WillExitWith(0)
				})

				Context("when TLS block is not configured", func() {
					It("calls mysql with the correct arguments", func() {
						By("calling mysql with the correct arguments for version checking", func() {
							expectedVersionCheckArgs := []string{
								fmt.Sprintf("--user=%s", username),
								fmt.Sprintf("--host=%s", host),
								fmt.Sprintf("--port=%d", port),
								"--skip-column-names",
								"--silent",
								`--execute=SELECT VERSION()`,
							}
							Expect(fakeMysqlClient57.Invocations()).To(HaveLen(2))
							Expect(fakeMysqlClient57.Invocations()[0].Args()).Should(ConsistOf(expectedVersionCheckArgs))
							Expect(fakeMysqlClient57.Invocations()[0].Env()).Should(HaveKeyWithValue("MYSQL_PWD", password))
						})

						By("calling mysql with the correct arguments for restoring", func() {
							expectedRestoreArgs := []interface{}{
								"-v",
								fmt.Sprintf("--user=%s", username),
								fmt.Sprintf("--host=%s", host),
								fmt.Sprintf("--port=%d", port),
								databaseName,
							}

							Expect(fakeMysqlClient57.Invocations()).To(HaveLen(2))
							Expect(fakeMysqlClient57.Invocations()[1].Args()).Should(ConsistOf(expectedRestoreArgs))
							Expect(fakeMysqlClient57.Invocations()[1].Stdin()).Should(ConsistOf("SOME BACKUP SQL"))
							Expect(fakeMysqlClient57.Invocations()[1].Env()).Should(HaveKeyWithValue("MYSQL_PWD", password))
						})

						By("succeeding", func() {
							Expect(session).Should(gexec.Exit(0))
						})
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

					It("calls mysql with the correct arguments", func() {
						By("calling mysql to detect the version", func() {
							Expect(fakeMysqlClient57.Invocations()).To(HaveLen(2))
							Expect(fakeMysqlClient57.Invocations()[0].Args()).Should(ConsistOf(
								fmt.Sprintf("--user=%s", username),
								fmt.Sprintf("--host=%s", host),
								fmt.Sprintf("--port=%d", port),
								HavePrefix("--ssl-ca="),
								"--ssl-mode=VERIFY_IDENTITY",
								"--skip-column-names",
								"--silent",
								`--execute=SELECT VERSION()`,
							))
							Expect(fakeMysqlClient57.Invocations()[0].Env()).Should(HaveKeyWithValue("MYSQL_PWD", password))
						})

						By("then calling mysql to restore", func() {
							expectedArgs := []interface{}{
								"-v",
								fmt.Sprintf("--user=%s", username),
								fmt.Sprintf("--host=%s", host),
								fmt.Sprintf("--port=%d", port),
								HavePrefix("--ssl-ca="),
								"--ssl-mode=VERIFY_IDENTITY",
								databaseName,
							}
							Expect(fakeMysqlClient57.Invocations()[1].Args()).Should(ConsistOf(expectedArgs))
							Expect(fakeMysqlClient57.Invocations()[1].Stdin()).Should(ConsistOf("SOME BACKUP SQL"))
						})
					})
				})

				Context("when TLS is configured with client cert and private key", func() {
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
									"ca": "A_CA_CERT",
									"certificate": "A_CLIENT_CERT",
									"private_key": "A_CLIENT_KEY"
								}
							}
						}`,
							username,
							password,
							host,
							port,
							databaseName))
					})

					It("calls mysql with the correct arguments", func() {
						By("calling mysql to detect the version", func() {
							Expect(fakeMysqlClient57.Invocations()).To(HaveLen(2))
							Expect(fakeMysqlClient57.Invocations()[0].Args()).Should(ConsistOf(
								fmt.Sprintf("--user=%s", username),
								fmt.Sprintf("--host=%s", host),
								fmt.Sprintf("--port=%d", port),
								HavePrefix("--ssl-ca="),
								HavePrefix("--ssl-cert="),
								HavePrefix("--ssl-key="),
								"--ssl-mode=VERIFY_IDENTITY",
								"--skip-column-names",
								"--silent",
								`--execute=SELECT VERSION()`,
							))
							Expect(fakeMysqlClient57.Invocations()[0].Env()).Should(
								HaveKeyWithValue("MYSQL_PWD", password))
						})

						By("then calling mysql to restore", func() {
							expectedArgs := []interface{}{
								fmt.Sprintf("--user=%s", username),
								fmt.Sprintf("--host=%s", host),
								fmt.Sprintf("--port=%d", port),
								HavePrefix("--ssl-ca="),
								HavePrefix("--ssl-cert="),
								HavePrefix("--ssl-key="),
								"--ssl-mode=VERIFY_IDENTITY",
								"-v",
								databaseName,
							}

							Expect(fakeMysqlClient57.Invocations()[1].Args()).Should(ConsistOf(expectedArgs))
							Expect(fakeMysqlClient57.Invocations()[1].Stdin()).Should(ConsistOf("SOME BACKUP SQL"))
						})
					})
				})
			})

			Context("when mysql fails", func() {
				BeforeEach(func() {
					fakeMysqlClient57.WhenCalledWith(
						fmt.Sprintf("--user=%s", username),
						fmt.Sprintf("--host=%s", host),
						fmt.Sprintf("--port=%d", port),
						"--skip-column-names",
						"--silent",
						`--execute=SELECT VERSION()`,
					).WillPrintToStdOut("MYSQL server version 5.7.27")
					fakeMysqlClient57.WhenCalled().WillExitWith(1)
				})

				It("also fails", func() {
					Expect(fakeMysqlClient57.Invocations()).To(HaveLen(2))
					Expect(fakeMysqlClient57.Invocations()[0].Env()).Should(HaveKeyWithValue("MYSQL_PWD", password))
					Eventually(session).Should(gexec.Exit(1))
				})
			})
		})
	})
	Context("mysql 8.0", func() {
		BeforeEach(func() {
			artifactFile = tempFilePath()
			fakeMysqlClient57.Reset()
			fakeMysqlDump80.Reset()
			fakeMysqlClient80.Reset()

			envVars["MYSQL_CLIENT_5_7_PATH"] = fakeMysqlClient57.Path
			envVars["MYSQL_CLIENT_8_0_PATH"] = fakeMysqlClient80.Path
			envVars["MYSQL_DUMP_8_0_PATH"] = fakeMysqlDump80.Path
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
					fakeMysqlClient57.WhenCalled().WillPrintToStdOut("MYSQL server version 8.0.27")
					fakeMysqlDump80.WhenCalled().WillExitWith(0)
				})

				It("calls mysql and mysqldump with the correct arguments", func() {
					By("calling mysql to detect the version", func() {
						Expect(fakeMysqlClient57.Invocations()).To(HaveLen(1))
						Expect(fakeMysqlClient57.Invocations()[0].Args()).Should(ConsistOf(
							fmt.Sprintf("--user=%s", username),
							fmt.Sprintf("--host=%s", host),
							fmt.Sprintf("--port=%d", port),
							"--skip-column-names",
							"--silent",
							`--execute=SELECT VERSION()`,
						))
						Expect(fakeMysqlClient57.Invocations()[0].Env()).Should(HaveKeyWithValue("MYSQL_PWD", password))
					})

					By("then calling dump", func() {
						Expect(fakeMysqlDump80.Invocations()).To(HaveLen(1))
						expectedArgs := []interface{}{
							fmt.Sprintf("--user=%s", username),
							fmt.Sprintf("--host=%s", host),
							fmt.Sprintf("--port=%d", port),
							"-v",
							"--set-gtid-purged=OFF",
							"--single-transaction",
							"--skip-add-locks",
							fmt.Sprintf("--result-file=%s", artifactFile),
							databaseName,
						}

						Expect(fakeMysqlDump80.Invocations()[0].Args()).Should(ConsistOf(expectedArgs))
						Expect(fakeMysqlDump80.Invocations()[0].Env()).Should(HaveKeyWithValue("MYSQL_PWD", password))
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
						expectedArgs := []interface{}{
							fmt.Sprintf("--user=%s", username),
							fmt.Sprintf("--host=%s", host),
							fmt.Sprintf("--port=%d", port),
							"-v",
							"--single-transaction",
							"--set-gtid-purged=OFF",
							"--skip-add-locks",
							fmt.Sprintf("--result-file=%s", artifactFile),
							databaseName,
							"table1",
							"table2",
							"table3",
						}

						Expect(fakeMysqlDump80.Invocations()[0].Args()).Should(ConsistOf(expectedArgs))
					})
				})

				Context("when TLS is configured with hostname verification turned off", func() {
					BeforeEach(func() {
						configFile = saveFile(fmt.Sprintf(`{
							"adapter":  "mysql",
							"username": "%s",
							"password": "%s",
							"host":     "%s",
							"port":     %d,
							"database": "%s",
							"tls": {
								"skip_host_verify": true,
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
							Expect(fakeMysqlClient57.Invocations()).To(HaveLen(1))
							Expect(fakeMysqlClient57.Invocations()[0].Args()).Should(ConsistOf(
								fmt.Sprintf("--user=%s", username),
								fmt.Sprintf("--host=%s", host),
								fmt.Sprintf("--port=%d", port),
								HavePrefix("--ssl-ca="),
								"--ssl-mode=VERIFY_CA",
								"--skip-column-names",
								"--silent",
								`--execute=SELECT VERSION()`,
							))
							Expect(fakeMysqlClient57.Invocations()[0].Env()).Should(
								HaveKeyWithValue("MYSQL_PWD", password))
						})

						By("then calling dump", func() {
							expectedArgs := []interface{}{
								fmt.Sprintf("--user=%s", username),
								fmt.Sprintf("--host=%s", host),
								fmt.Sprintf("--port=%d", port),
								HavePrefix("--ssl-ca="),
								"--ssl-mode=VERIFY_CA",
								"--set-gtid-purged=OFF",
								"-v",
								"--single-transaction",
								"--skip-add-locks",
								fmt.Sprintf("--result-file=%s", artifactFile),
								databaseName,
							}

							Expect(fakeMysqlDump80.Invocations()[0].Args()).Should(ConsistOf(expectedArgs))
						})
					})
				})

				Context("when TLS is configured with client cert and private key", func() {
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
									"ca": "A_CA_CERT",
									"certificate": "A_CLIENT_CERT",
									"private_key": "A_CLIENT_KEY"
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
							Expect(fakeMysqlClient57.Invocations()).To(HaveLen(1))
							Expect(fakeMysqlClient57.Invocations()[0].Args()).Should(ConsistOf(
								fmt.Sprintf("--user=%s", username),
								fmt.Sprintf("--host=%s", host),
								fmt.Sprintf("--port=%d", port),
								HavePrefix("--ssl-ca="),
								HavePrefix("--ssl-cert="),
								HavePrefix("--ssl-key="),
								"--ssl-mode=VERIFY_IDENTITY",
								"--skip-column-names",
								"--silent",
								`--execute=SELECT VERSION()`,
							))
							Expect(fakeMysqlClient57.Invocations()[0].Env()).Should(
								HaveKeyWithValue("MYSQL_PWD", password))
						})

						By("then calling dump", func() {
							expectedArgs := []interface{}{
								fmt.Sprintf("--user=%s", username),
								fmt.Sprintf("--host=%s", host),
								fmt.Sprintf("--port=%d", port),
								HavePrefix("--ssl-ca="),
								HavePrefix("--ssl-cert="),
								HavePrefix("--ssl-key="),
								"--ssl-mode=VERIFY_IDENTITY",
								"-v",
								"--single-transaction",
								"--skip-add-locks",
								"--set-gtid-purged=OFF",
								fmt.Sprintf("--result-file=%s", artifactFile),
								databaseName,
							}

							Expect(fakeMysqlDump80.Invocations()[0].Args()).Should(ConsistOf(expectedArgs))
						})
					})
				})
			})

			Context("when version detection fails", func() {
				BeforeEach(func() {
					fakeMysqlClient57.WhenCalledWith(
						fmt.Sprintf("--user=%s", username),
						fmt.Sprintf("--host=%s", host),
						fmt.Sprintf("--port=%d", port),
						"--skip-column-names",
						"--silent",
						`--execute=SELECT VERSION()`,
					).WillExitWith(1).WillPrintToStdErr("VERSION DETECTION FAILED!")
				})

				It("also fails", func() {
					Expect(fakeMysqlClient57.Invocations()[0].Env()).Should(HaveKeyWithValue("MYSQL_PWD", password))
					Eventually(session).Should(gexec.Exit(1))
					Expect(session.Err).To(gbytes.Say("VERSION DETECTION FAILED!"))
				})
			})

			Context("when mysqldump fails", func() {
				BeforeEach(func() {
					fakeMysqlClient57.WhenCalledWith(
						fmt.Sprintf("--user=%s", username),
						fmt.Sprintf("--host=%s", host),
						fmt.Sprintf("--port=%d", port),
						"--skip-column-names",
						"--silent",
						`--execute=SELECT VERSION()`,
					).WillPrintToStdOut("MYSQL server version 8.0.38")
					fakeMysqlDump80.WhenCalled().WillExitWith(1)
				})

				It("also fails", func() {
					Expect(fakeMysqlClient57.Invocations()[0].Env()).Should(HaveKeyWithValue("MYSQL_PWD", password))
					Eventually(session).Should(gexec.Exit(1))
				})
			})

			Context("when the server has an unsupported mysql major version", func() {
				BeforeEach(func() {
					fakeMysqlClient57.WhenCalledWith(
						fmt.Sprintf("--user=%s", username),
						fmt.Sprintf("--host=%s", host),
						fmt.Sprintf("--port=%d", port),
						"--skip-column-names",
						"--silent",
						`--execute=SELECT VERSION()`,
					).WillPrintToStdOut("MYSQL server version 4.7.20")
				})

				It("fails because of a version mismatch", func() {
					Expect(fakeMysqlClient57.Invocations()[0].Env()).Should(HaveKeyWithValue("MYSQL_PWD", password))
					Expect(session).Should(gexec.Exit(1))
					Expect(string(session.Err.Contents())).Should(ContainSubstring(
						"unsupported version of mysql: 4.7"),
					)
				})
			})
			Context("when the server has a reviously supported mysql version", func() {
				BeforeEach(func() {
					fakeMysqlClient57.WhenCalledWith(
						fmt.Sprintf("--user=%s", username),
						fmt.Sprintf("--host=%s", host),
						fmt.Sprintf("--port=%d", port),
						"--skip-column-names",
						"--silent",
						`--execute=SELECT VERSION()`,
					).WillPrintToStdOut("MYSQL server version 5.6.20")
				})

				It("fails because of a version mismatch", func() {
					Expect(fakeMysqlClient57.Invocations()[0].Env()).Should(HaveKeyWithValue("MYSQL_PWD", password))
					Expect(session).Should(gexec.Exit(1))
					Expect(string(session.Err.Contents())).Should(ContainSubstring(
						"unsupported version of mysql: 5.6"),
					)
				})
			})

			Context("when the server has an unsupported mysql minor version", func() {
				BeforeEach(func() {
					fakeMysqlClient57.WhenCalledWith(
						fmt.Sprintf("--user=%s", username),
						fmt.Sprintf("--host=%s", host),
						fmt.Sprintf("--port=%d", port),
						"--skip-column-names",
						"--silent",
						`--execute=SELECT VERSION()`,
					).WillPrintToStdOut("MYSQL server version 5.9.20")
				})

				It("fails because of a version mismatch", func() {
					Expect(fakeMysqlClient57.Invocations()[0].Env()).Should(HaveKeyWithValue("MYSQL_PWD", password))
					Expect(session).Should(gexec.Exit(1))
					Expect(string(session.Err.Contents())).Should(ContainSubstring(
						"unsupported version of mysql: 5.9"),
					)
				})
			})

			Context("when the server has a supported mysql minor version with a different patch to the packaged utility", func() {
				BeforeEach(func() {
					fakeMysqlDump80.WhenCalled().WillExitWith(0)
					fakeMysqlClient57.WhenCalledWith(
						fmt.Sprintf("--user=%s", username),
						fmt.Sprintf("--host=%s", host),
						fmt.Sprintf("--port=%d", port),
						"--skip-column-names",
						"--silent",
						`--execute=SELECT VERSION()`,
					).WillPrintToStdOut("MYSQL server version 8.0.666")
				})

				It("succeeds despite different patch versions", func() {
					Expect(fakeMysqlClient57.Invocations()[0].Env()).Should(HaveKeyWithValue("MYSQL_PWD", password))
					Expect(fakeMysqlDump80.Invocations()).Should(HaveLen(1))
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
				err := os.WriteFile(artifactFile, []byte("SOME BACKUP SQL"), 0644)
				Expect(err).ToNot(HaveOccurred())

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
					fakeMysqlClient57.WhenCalled().WillPrintToStdOut("MYSQL server version 8.0.27")
					fakeMysqlClient80.WhenCalled().WillExitWith(0)
				})

				Context("when TLS block is not configured", func() {
					It("calls mysql with the correct arguments", func() {
						By("calling mysql with the correct arguments for version checking", func() {
							expectedVersionCheckArgs := []string{
								fmt.Sprintf("--user=%s", username),
								fmt.Sprintf("--host=%s", host),
								fmt.Sprintf("--port=%d", port),
								"--skip-column-names",
								"--silent",
								`--execute=SELECT VERSION()`,
							}
							Expect(fakeMysqlClient57.Invocations()).To(HaveLen(1))
							Expect(fakeMysqlClient57.Invocations()[0].Args()).Should(ConsistOf(expectedVersionCheckArgs))
							Expect(fakeMysqlClient57.Invocations()[0].Env()).Should(HaveKeyWithValue("MYSQL_PWD", password))
						})

						By("calling mysql with the correct arguments for restoring", func() {
							expectedRestoreArgs := []interface{}{
								"-v",
								fmt.Sprintf("--user=%s", username),
								fmt.Sprintf("--host=%s", host),
								fmt.Sprintf("--port=%d", port),
								databaseName,
							}

							Expect(fakeMysqlClient80.Invocations()).To(HaveLen(1))
							Expect(fakeMysqlClient80.Invocations()[0].Args()).Should(ConsistOf(expectedRestoreArgs))
							Expect(fakeMysqlClient80.Invocations()[0].Stdin()).Should(ConsistOf("SOME BACKUP SQL"))
							Expect(fakeMysqlClient80.Invocations()[0].Env()).Should(HaveKeyWithValue("MYSQL_PWD", password))
						})

						By("succeeding", func() {
							Expect(session).Should(gexec.Exit(0))
						})
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

					It("calls mysql with the correct arguments", func() {
						By("calling mysql to detect the version", func() {
							Expect(fakeMysqlClient57.Invocations()).To(HaveLen(1))
							Expect(fakeMysqlClient57.Invocations()[0].Args()).Should(ConsistOf(
								fmt.Sprintf("--user=%s", username),
								fmt.Sprintf("--host=%s", host),
								fmt.Sprintf("--port=%d", port),
								HavePrefix("--ssl-ca="),
								"--ssl-mode=VERIFY_IDENTITY",
								"--skip-column-names",
								"--silent",
								`--execute=SELECT VERSION()`,
							))
							Expect(fakeMysqlClient57.Invocations()[0].Env()).Should(HaveKeyWithValue("MYSQL_PWD", password))
						})

						By("then calling mysql to restore", func() {
							expectedArgs := []interface{}{
								"-v",
								fmt.Sprintf("--user=%s", username),
								fmt.Sprintf("--host=%s", host),
								fmt.Sprintf("--port=%d", port),
								HavePrefix("--ssl-ca="),
								"--ssl-mode=VERIFY_IDENTITY",
								databaseName,
							}

							Expect(fakeMysqlClient80.Invocations()).To(HaveLen(1))
							Expect(fakeMysqlClient80.Invocations()[0].Args()).Should(ConsistOf(expectedArgs))
							Expect(fakeMysqlClient80.Invocations()[0].Stdin()).Should(ConsistOf("SOME BACKUP SQL"))
						})
					})
				})

				Context("when TLS is configured with client cert and private key", func() {
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
									"ca": "A_CA_CERT",
									"certificate": "A_CLIENT_CERT",
									"private_key": "A_CLIENT_KEY"
								}
							}
						}`,
							username,
							password,
							host,
							port,
							databaseName))
					})

					It("calls mysql with the correct arguments", func() {
						By("calling mysql to detect the version", func() {
							Expect(fakeMysqlClient57.Invocations()).To(HaveLen(1))
							Expect(fakeMysqlClient57.Invocations()[0].Args()).Should(ConsistOf(
								fmt.Sprintf("--user=%s", username),
								fmt.Sprintf("--host=%s", host),
								fmt.Sprintf("--port=%d", port),
								HavePrefix("--ssl-ca="),
								HavePrefix("--ssl-cert="),
								HavePrefix("--ssl-key="),
								"--ssl-mode=VERIFY_IDENTITY",
								"--skip-column-names",
								"--silent",
								`--execute=SELECT VERSION()`,
							))
							Expect(fakeMysqlClient57.Invocations()[0].Env()).Should(
								HaveKeyWithValue("MYSQL_PWD", password))
						})

						By("then calling mysql to restore", func() {
							expectedArgs := []interface{}{
								fmt.Sprintf("--user=%s", username),
								fmt.Sprintf("--host=%s", host),
								fmt.Sprintf("--port=%d", port),
								HavePrefix("--ssl-ca="),
								HavePrefix("--ssl-cert="),
								HavePrefix("--ssl-key="),
								"--ssl-mode=VERIFY_IDENTITY",
								"-v",
								databaseName,
							}

							Expect(fakeMysqlClient80.Invocations()).To(HaveLen(1))
							Expect(fakeMysqlClient80.Invocations()[0].Args()).Should(ConsistOf(expectedArgs))
							Expect(fakeMysqlClient80.Invocations()[0].Stdin()).Should(ConsistOf("SOME BACKUP SQL"))
						})
					})
				})
			})

			Context("when mysql fails", func() {
				BeforeEach(func() {
					fakeMysqlClient57.WhenCalledWith(
						fmt.Sprintf("--user=%s", username),
						fmt.Sprintf("--host=%s", host),
						fmt.Sprintf("--port=%d", port),
						"--skip-column-names",
						"--silent",
						`--execute=SELECT VERSION()`,
					).WillPrintToStdOut("MYSQL server version 8.0.27")
					fakeMysqlClient80.WhenCalled().WillExitWith(1)
				})

				It("also fails", func() {
					Expect(fakeMysqlClient57.Invocations()).To(HaveLen(1))
					Expect(fakeMysqlClient80.Invocations()).To(HaveLen(1))
					Expect(fakeMysqlClient57.Invocations()[0].Env()).Should(HaveKeyWithValue("MYSQL_PWD", password))
					Eventually(session).Should(gexec.Exit(1))
				})
			})
		})
	})

})

var _ = Describe("MariaDB", func() {
	var session *gexec.Session
	var username = "testuser"
	var host = "127.0.0.1"
	var port = 1234
	var databaseName = "mycooldb"
	var password = "password"
	var artifactFile string
	var err error
	var configFile *os.File

	Context("mariadb 10.1", func() {
		BeforeEach(func() {
			artifactFile = tempFilePath()
			fakeMariaDBDump.Reset()
			fakeMysqlClient57.Reset()

			envVars["MYSQL_CLIENT_5_7_PATH"] = fakeMysqlClient57.Path
			envVars["MARIADB_DUMP_PATH"] = fakeMariaDBDump.Path
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
					fakeMysqlClient57.WhenCalled().WillPrintToStdOut("10.1.34-MariaDB")
					fakeMariaDBDump.WhenCalled().WillExitWith(0)
				})

				It("calls mysqldump with the correct arguments", func() {
					Expect(fakeMysqlClient57.Invocations()).To(HaveLen(1))
					Expect(fakeMariaDBDump.Invocations()).To(HaveLen(1))
					expectedArgs := []interface{}{
						fmt.Sprintf("--user=%s", username),
						fmt.Sprintf("--host=%s", host),
						fmt.Sprintf("--port=%d", port),
						HavePrefix("--ssl-cipher="),
						"-v",
						"--single-transaction",
						"--skip-add-locks",
						fmt.Sprintf("--result-file=%s", artifactFile),
						databaseName,
					}

					Expect(fakeMariaDBDump.Invocations()[0].Args()).Should(ConsistOf(expectedArgs))
					Expect(fakeMariaDBDump.Invocations()[0].Env()).Should(HaveKeyWithValue("MYSQL_PWD", password))

					Expect(session).Should(gexec.Exit(0))
				})

				Context("when TLS is configured with hostname verification turned off", func() {
					BeforeEach(func() {
						configFile = saveFile(fmt.Sprintf(`{
							"adapter":  "mysql",
							"username": "%s",
							"password": "%s",
							"host":     "%s",
							"port":     %d,
							"database": "%s",
							"tls": {
								"skip_host_verify": true,
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

					It("calls mysqldump with the correct arguments", func() {
						expectedArgs := []interface{}{
							fmt.Sprintf("--user=%s", username),
							fmt.Sprintf("--host=%s", host),
							fmt.Sprintf("--port=%d", port),
							HavePrefix("--ssl-ca="),
							"-v",
							"--single-transaction",
							"--skip-add-locks",
							fmt.Sprintf("--result-file=%s", artifactFile),
							databaseName,
						}

						Expect(fakeMariaDBDump.Invocations()[0].Args()).Should(ConsistOf(expectedArgs))
					})
				})

				Context("when TLS is configured with client cert and private key", func() {
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
									"ca": "A_CA_CERT",
									"certificate": "A_CLIENT_CERT",
									"private_key": "A_CLIENT_KEY"
								}
							}
						}`,
							username,
							password,
							host,
							port,
							databaseName))
					})

					It("calls mysqldump with the correct arguments", func() {
						expectedArgs := []interface{}{
							fmt.Sprintf("--user=%s", username),
							fmt.Sprintf("--host=%s", host),
							fmt.Sprintf("--port=%d", port),
							HavePrefix("--ssl-ca="),
							HavePrefix("--ssl-cert="),
							HavePrefix("--ssl-key="),
							"--ssl-verify-server-cert",
							"-v",
							"--single-transaction",
							"--skip-add-locks",
							fmt.Sprintf("--result-file=%s", artifactFile),
							databaseName,
						}

						Expect(fakeMariaDBDump.Invocations()[0].Args()).Should(ConsistOf(expectedArgs))
					})
				})
			})
		})
	})
})
