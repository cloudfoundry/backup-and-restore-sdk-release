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

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/pivotal-cf-experimental/go-binmock"
)

var _ = Describe("Postgres", func() {
	var fakePgDump94 *binmock.Mock
	var fakePgDump96 *binmock.Mock
	var fakePsql *binmock.Mock
	var clientIsConnectedTo94 *binmock.Mock
	var clientIsConnectedTo96 *binmock.Mock
	var session *gexec.Session
	var username = "testuser"
	var host = "127.0.0.1"
	var port = 1234
	var databaseName = "mycooldb"
	var password = "password"
	var artifactFile string
	var compiledSDKPath string
	var err error
	var configFile *os.File

	BeforeEach(func() {
		compiledSDKPath, err = gexec.Build(
			"github.com/cloudfoundry-incubator/database-backup-and-restore/cmd/database-backup-restore")
		Expect(err).NotTo(HaveOccurred())

		artifactFile = tempFilePath()
	})

	Context("backup", func() {

		BeforeEach(func() {
			configFile = buildConfigFile(Config{
				Adapter:  "postgres",
				Username: username,
				Password: password,
				Host:     host,
				Port:     port,
				Database: databaseName,
			})

			clientIsConnectedTo96 = binmock.NewBinMock(Fail)
			clientIsConnectedTo96.WhenCalled().WillPrintToStdOut(
				" PostgreSQL 9.6.3 on x86_64-pc-linux-gnu, compiled by gcc " + "" +
					"(Ubuntu 4.8.4-2ubuntu1~14.04.3) 4.8.4, 64-bit").
				WillExitWith(0)

			clientIsConnectedTo94 = binmock.NewBinMock(Fail)
			clientIsConnectedTo94.WhenCalled().WillPrintToStdOut(
				" PostgreSQL 9.4.9 on x86_64-unknown-linux-gnu, compiled by gcc " +
					"(Ubuntu 4.8.4-2ubuntu1~14.04.3) 4.8.4, 64-bit").
				WillExitWith(0)

			fakePgDump94 = binmock.NewBinMock(Fail)
			fakePgDump94.WhenCalled().WillExitWith(0)
			fakePgDump96 = binmock.NewBinMock(Fail)
			fakePgDump96.WhenCalled().WillExitWith(0)
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

		Context("PG_CLIENT_PATH env var is missing", func() {
			BeforeEach(func() {
				delete(envVars, "PG_CLIENT_PATH")
			})

			It("raises an appropriate error", func() {
				Expect(session.Err).To(gbytes.Say("PG_CLIENT_PATH must be set"))
			})
		})

		Context("PG_CLIENT_PATH env var is set", func() {
			Context("Postgres database server is version 9.4", func() {

				BeforeEach(func() {
					envVars["PG_CLIENT_PATH"] = clientIsConnectedTo94.Path
				})

				Context("PG_DUMP_9_4 env var is missing", func() {
					BeforeEach(func() {
						delete(envVars, "PG_DUMP_9_4_PATH")
					})

					It("raises an appropriate error", func() {
						Expect(session.Err).To(gbytes.Say("PG_DUMP_9_4_PATH must be set"))
					})
				})

				Context("PG_DUMP_9_4 env var is set", func() {
					BeforeEach(func() {
						envVars["PG_DUMP_9_4_PATH"] = fakePgDump94.Path
					})
					It("takes a backup", func() {
						By("getting the server version", func() {
							expectedArgs := []string{
								"--tuples-only",
								fmt.Sprintf("--username=%s", username),
								fmt.Sprintf("--host=%s", host),
								fmt.Sprintf("--port=%d", port),
								databaseName,
								`--command=SELECT VERSION()`,
							}

							Expect(clientIsConnectedTo94.Invocations()).To(HaveLen(1))
							Expect(clientIsConnectedTo94.Invocations()[0].Args()).Should(ConsistOf(expectedArgs))
							Expect(clientIsConnectedTo94.Invocations()[0].Env()).Should(HaveKeyWithValue("PGPASSWORD", password))
						})

						By("dumping the database with the correct dump binary", func() {
							expectedArgs := []string{
								"-v",
								fmt.Sprintf("--user=%s", username),
								fmt.Sprintf("--host=%s", host),
								fmt.Sprintf("--port=%d", port),
								"--format=custom",
								fmt.Sprintf("--file=%s", artifactFile),
								databaseName,
							}

							Expect(fakePgDump94.Invocations()).To(HaveLen(1))
							Expect(fakePgDump94.Invocations()[0].Args()).Should(ConsistOf(expectedArgs))
							Expect(fakePgDump94.Invocations()[0].Env()).Should(HaveKeyWithValue("PGPASSWORD", password))
						})

						By("not invoking the dump binary for a different version", func() {
							Expect(fakePgDump96.Invocations()).To(HaveLen(0))
						})

						Expect(session).Should(gexec.Exit(0))
					})

					Context("when 'tables' are specified in the configFile", func() {
						BeforeEach(func() {
							configFile = buildConfigFile(Config{
								Adapter:  "postgres",
								Username: username,
								Password: password,
								Host:     host,
								Port:     port,
								Database: databaseName,
								Tables:   []string{"table1", "table2", "table3"},
							})
							clientIsConnectedTo94.WhenCalled().WillPrintToStdOut(
								" table1 \n table2 \n table3 \n\n\n").
								WillExitWith(0)
						})

						It("backs up the specified tables", func() {
							By("checking if the tables exist", func() {
								expectedArgs := []string{
									"--tuples-only",
									fmt.Sprintf("--username=%s", username),
									fmt.Sprintf("--host=%s", host),
									fmt.Sprintf("--port=%d", port),
									databaseName,
									`--command=SELECT table_name FROM information_schema.tables WHERE table_type='BASE TABLE' AND table_schema='public';`,
								}

								Expect(clientIsConnectedTo94.Invocations()).To(HaveLen(2))
								Expect(clientIsConnectedTo94.Invocations()[1].Args()).Should(ConsistOf(expectedArgs))
								Expect(clientIsConnectedTo94.Invocations()[1].Env()).Should(HaveKeyWithValue("PGPASSWORD", password))
							})

							By("calling pg_dump with the correct arguments", func() {
								expectedArgs := []string{
									"-v",
									fmt.Sprintf("--user=%s", username),
									fmt.Sprintf("--host=%s", host),
									fmt.Sprintf("--port=%d", port),
									"--format=custom",
									fmt.Sprintf("--file=%s", artifactFile),
									databaseName,
									"-t", "table1",
									"-t", "table2",
									"-t", "table3",
								}

								Expect(fakePgDump94.Invocations()[0].Args()).Should(ConsistOf(expectedArgs))
							})

							Expect(session).Should(gexec.Exit(0))
						})
					})

					Context("when missing 'tables' are specified in the configFile", func() {
						BeforeEach(func() {
							configFile = buildConfigFile(Config{
								Adapter:  "postgres",
								Username: username,
								Password: password,
								Host:     host,
								Port:     port,
								Database: databaseName,
								Tables:   []string{"table1", "table2", "table3"},
							})
							clientIsConnectedTo94.WhenCalled().WillPrintToStdOut(
								" table1 \n table2 \n\n\n").
								WillExitWith(0)
						})

						It("fails", func() {
							By("checking if the tables exist", func() {
								expectedArgs := []string{
									"--tuples-only",
									fmt.Sprintf("--username=%s", username),
									fmt.Sprintf("--host=%s", host),
									fmt.Sprintf("--port=%d", port),
									databaseName,
									`--command=SELECT table_name FROM information_schema.tables WHERE table_type='BASE TABLE' AND table_schema='public';`,
								}

								Expect(clientIsConnectedTo94.Invocations()).To(HaveLen(2))
								Expect(clientIsConnectedTo94.Invocations()[1].Args()).Should(ConsistOf(expectedArgs))
								Expect(clientIsConnectedTo94.Invocations()[1].Env()).Should(HaveKeyWithValue("PGPASSWORD", password))
							})

							By("exiting with a helpful error message", func() {
								Expect(session).Should(gexec.Exit(1))
								Expect(session.Err).Should(gbytes.Say(`can't find specified table\(s\): table3`))
							})
						})
					})

					Context("and pg_dump fails", func() {
						BeforeEach(func() {
							fakePgDump94 = binmock.NewBinMock(Fail)
							fakePgDump94.WhenCalled().WillExitWith(1)

							envVars["PG_DUMP_9_4_PATH"] = fakePgDump94.Path
						})

						It("also fails", func() {
							Eventually(session).Should(gexec.Exit(1))
						})
					})
				})

			})

			Context("Postgres database server is version 9.6", func() {
				BeforeEach(func() {
					envVars["PG_CLIENT_PATH"] = clientIsConnectedTo96.Path
				})

				Context("PG_DUMP_9_6 env var is missing", func() {
					BeforeEach(func() {
						delete(envVars, "PG_DUMP_9_6_PATH")
					})

					It("raises an appropriate error", func() {
						Expect(session.Err).To(gbytes.Say("PG_DUMP_9_6_PATH must be set"))
					})
				})

				Context("PG_DUMP_9_6 env var is set", func() {
					BeforeEach(func() {
						envVars["PG_DUMP_9_6_PATH"] = fakePgDump96.Path
					})

					It("takes a backup", func() {
						By("getting the server version", func() {
							expectedArgs := []string{
								"--tuples-only",
								fmt.Sprintf("--username=%s", username),
								fmt.Sprintf("--host=%s", host),
								fmt.Sprintf("--port=%d", port),
								databaseName,
								`--command=SELECT VERSION()`,
							}

							Expect(clientIsConnectedTo96.Invocations()).To(HaveLen(1))
							Expect(clientIsConnectedTo96.Invocations()[0].Args()).Should(ConsistOf(expectedArgs))
							Expect(clientIsConnectedTo96.Invocations()[0].Env()).Should(HaveKeyWithValue("PGPASSWORD", password))
						})

						By("dumping the database with the correct dump binary", func() {
							expectedArgs := []string{
								"-v",
								fmt.Sprintf("--user=%s", username),
								fmt.Sprintf("--host=%s", host),
								fmt.Sprintf("--port=%d", port),
								"--format=custom",
								fmt.Sprintf("--file=%s", artifactFile),
								databaseName,
							}

							Expect(fakePgDump96.Invocations()).To(HaveLen(1))
							Expect(fakePgDump96.Invocations()[0].Args()).Should(ConsistOf(expectedArgs))
							Expect(fakePgDump96.Invocations()[0].Env()).Should(HaveKeyWithValue("PGPASSWORD", password))
						})

						By("not invoking the dump binary for a different version", func() {
							Expect(fakePgDump94.Invocations()).To(HaveLen(0))
						})

						Expect(session).Should(gexec.Exit(0))
					})

					Context("when 'tables' are specified in the configFile", func() {
						BeforeEach(func() {
							configFile = buildConfigFile(Config{
								Adapter:  "postgres",
								Username: username,
								Password: password,
								Host:     host,
								Port:     port,
								Database: databaseName,
								Tables:   []string{"table1", "table2", "table3"},
							})
							clientIsConnectedTo96.WhenCalled().WillPrintToStdOut(
								" table1 \n table2 \n table3 \n\n\n").
								WillExitWith(0)
						})

						It("backs up the specified tables", func() {
							By("checking if the tables exist", func() {
								expectedArgs := []string{
									"--tuples-only",
									fmt.Sprintf("--username=%s", username),
									fmt.Sprintf("--host=%s", host),
									fmt.Sprintf("--port=%d", port),
									databaseName,
									`--command=SELECT table_name FROM information_schema.tables WHERE table_type='BASE TABLE' AND table_schema='public';`,
								}

								Expect(clientIsConnectedTo96.Invocations()).To(HaveLen(2))
								Expect(clientIsConnectedTo96.Invocations()[1].Args()).Should(ConsistOf(expectedArgs))
								Expect(clientIsConnectedTo96.Invocations()[1].Env()).Should(HaveKeyWithValue("PGPASSWORD", password))
							})

							By("calling pg_dump with the correct arguments", func() {
								expectedArgs := []string{
									"-v",
									fmt.Sprintf("--user=%s", username),
									fmt.Sprintf("--host=%s", host),
									fmt.Sprintf("--port=%d", port),
									"--format=custom",
									fmt.Sprintf("--file=%s", artifactFile),
									databaseName,
									"-t", "table1",
									"-t", "table2",
									"-t", "table3",
								}

								Expect(fakePgDump96.Invocations()[0].Args()).Should(ConsistOf(expectedArgs))
							})

							Expect(session).Should(gexec.Exit(0))
						})
					})

					Context("when missing 'tables' are specified in the configFile", func() {
						BeforeEach(func() {
							configFile = buildConfigFile(Config{
								Adapter:  "postgres",
								Username: username,
								Password: password,
								Host:     host,
								Port:     port,
								Database: databaseName,
								Tables:   []string{"table1", "table2", "table3"},
							})
							clientIsConnectedTo96.WhenCalled().WillPrintToStdOut(
								" table1 \n table2 \n\n\n").
								WillExitWith(0)
						})

						It("fails", func() {
							By("checking if the tables exist", func() {
								expectedArgs := []string{
									"--tuples-only",
									fmt.Sprintf("--username=%s", username),
									fmt.Sprintf("--host=%s", host),
									fmt.Sprintf("--port=%d", port),
									databaseName,
									`--command=SELECT table_name FROM information_schema.tables WHERE table_type='BASE TABLE' AND table_schema='public';`,
								}

								Expect(clientIsConnectedTo96.Invocations()).To(HaveLen(2))
								Expect(clientIsConnectedTo96.Invocations()[1].Args()).Should(ConsistOf(expectedArgs))
								Expect(clientIsConnectedTo96.Invocations()[1].Env()).Should(HaveKeyWithValue("PGPASSWORD", password))
							})

							By("exiting with a helpful error message", func() {
								Expect(session).Should(gexec.Exit(1))
								Expect(session.Err).Should(gbytes.Say(`can't find specified table\(s\): table3`))
							})
						})
					})
				})

				Context("and pg_dump fails", func() {
					BeforeEach(func() {
						fakePgDump96 = binmock.NewBinMock(Fail)
						fakePgDump96.WhenCalled().WillExitWith(1)

						envVars["PG_DUMP_9_6_PATH"] = fakePgDump96.Path
					})

					It("also fails", func() {
						Eventually(session).Should(gexec.Exit(1))
					})
				})
			})
		})

	})

	Context("restore", func() {
		BeforeEach(func() {
			configFile = buildConfigFile(Config{
				Adapter:  "postgres",
				Username: username,
				Password: password,
				Host:     host,
				Port:     port,
				Database: databaseName,
			})

			fakePsql = binmock.NewBinMock(Fail)
			fakePsql.WhenCalled().WillExitWith(0)

		})

		JustBeforeEach(func() {
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

		Context("PG_RESTORE_9_4_PATH env var is missing", func() {
			BeforeEach(func() {
				delete(envVars, "PG_RESTORE_9_4_PATH")
			})

			It("raises an appropriate error", func() {
				Expect(session.Err).To(gbytes.Say("PG_RESTORE_9_4_PATH must be set"))
			})
		})

		Context("PG_RESTORE_9_4_PATH is set", func() {
			BeforeEach(func() {
				envVars["PG_RESTORE_9_4_PATH"] = fakePsql.Path
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

				Expect(fakePsql.Invocations()).To(HaveLen(1))
				Expect(fakePsql.Invocations()[0].Args()).Should(ConsistOf(expectedArgs))
				Expect(fakePsql.Invocations()[0].Env()).Should(HaveKeyWithValue("PGPASSWORD", password))
			})

			It("succeeds", func() {
				Expect(session).Should(gexec.Exit(0))
			})

			Context("and pg_restore fails", func() {
				BeforeEach(func() {
					fakePsql = binmock.NewBinMock(Fail)
					fakePsql.WhenCalled().WillExitWith(1)
					envVars["PG_RESTORE_9_4_PATH"] = fakePsql.Path
				})

				It("also fails", func() {
					Eventually(session).Should(gexec.Exit(1))
				})
			})
		})
	})
})
