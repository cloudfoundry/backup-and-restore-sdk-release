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
)

var _ = Describe("Postgres", func() {
	var session *gexec.Session
	var username = "testuser"
	var host = "127.0.0.1"
	var port = 1234
	var databaseName = "mycooldb"
	var password = "password"
	var artifactFile string
	var configFile *os.File

	BeforeEach(func() {
		artifactFile = tempFilePath()

		fakePgClient.Reset()
		fakePgDump10.Reset()
		fakePgDump11.Reset()
		fakePgDump13.Reset()
		fakePgDump15.Reset()
		fakePgRestore10.Reset()
		fakePgRestore11.Reset()
		fakePgRestore13.Reset()
		fakePgRestore15.Reset()

		envVars["PG_CLIENT_PATH"] = fakePgClient.Path
		envVars["PG_DUMP_10_PATH"] = fakePgDump10.Path
		envVars["PG_DUMP_11_PATH"] = fakePgDump11.Path
		envVars["PG_DUMP_13_PATH"] = fakePgDump13.Path
		envVars["PG_DUMP_15_PATH"] = fakePgDump15.Path
		envVars["PG_RESTORE_10_PATH"] = fakePgRestore10.Path
		envVars["PG_RESTORE_11_PATH"] = fakePgRestore11.Path
		envVars["PG_RESTORE_13_PATH"] = fakePgRestore13.Path
		envVars["PG_RESTORE_15_PATH"] = fakePgRestore15.Path

		configFile = saveFile(fmt.Sprintf(`{
				"adapter":  "postgres",
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

	Context("backup", func() {
		JustBeforeEach(func() {
			session = run(compiledSDKPath, envVars,
				"--artifact-file", artifactFile,
				"--config", configFile.Name(),
				"--backup",
			)
		})

		Context("Postgres database server is version 10", func() {
			BeforeEach(func() {
				fakePgClient.WhenCalled().WillPrintToStdOut(
					" PostgreSQL 10.6 on x86_64-pc-linux-gnu, compiled by gcc " +
						"(Ubuntu 5.4.0-6ubuntu1~16.04.11) 5.4.0 20160609, 64-bit").
					WillExitWith(0)
			})

			Context("when pg_dump succeeds", func() {
				BeforeEach(func() {
					fakePgDump10.WhenCalled().WillExitWith(0)
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

						Expect(fakePgClient.Invocations()).To(HaveLen(1))
						Expect(fakePgClient.Invocations()[0].Args()).Should(ConsistOf(expectedArgs))
						Expect(fakePgClient.Invocations()[0].Env()).Should(HaveKeyWithValue("PGPASSWORD", password))
					})

					By("dumping the database with the correct dump binary", func() {
						expectedArgs := []string{
							"--verbose",
							fmt.Sprintf("--username=%s", username),
							fmt.Sprintf("--host=%s", host),
							fmt.Sprintf("--port=%d", port),
							"--format=custom",
							fmt.Sprintf("--file=%s", artifactFile),
							databaseName,
						}

						Expect(fakePgDump10.Invocations()).To(HaveLen(1))
						Expect(fakePgDump10.Invocations()[0].Args()).Should(ConsistOf(expectedArgs))
						Expect(fakePgDump10.Invocations()[0].Env()).Should(HaveKeyWithValue("PGPASSWORD", password))
					})

					By("not invoking the dump binary for a different version", func() {
						Expect(fakePgDump11.Invocations()).To(HaveLen(0))
					})

					Expect(session).Should(gexec.Exit(0))
				})

				Context("when 'tables' are specified in the configFile", func() {
					BeforeEach(func() {
						configFile = saveFile(fmt.Sprintf(`{
								"adapter":  "postgres",
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
						fakePgClient.WhenCalled().WillPrintToStdOut(
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

							Expect(fakePgClient.Invocations()).To(HaveLen(2))
							Expect(fakePgClient.Invocations()[1].Args()).Should(ConsistOf(expectedArgs))
							Expect(fakePgClient.Invocations()[1].Env()).Should(HaveKeyWithValue("PGPASSWORD", password))
						})

						By("calling pg_dump with the correct arguments", func() {
							expectedArgs := []string{
								"--verbose",
								fmt.Sprintf("--username=%s", username),
								fmt.Sprintf("--host=%s", host),
								fmt.Sprintf("--port=%d", port),
								"--format=custom",
								fmt.Sprintf("--file=%s", artifactFile),
								databaseName,
								"-t", "table1",
								"-t", "table2",
								"-t", "table3",
							}

							Expect(fakePgDump10.Invocations()[0].Args()).Should(ConsistOf(expectedArgs))
						})

						Expect(session).Should(gexec.Exit(0))
					})
				})

				Context("when missing 'tables' are specified in the configFile", func() {
					BeforeEach(func() {
						configFile = saveFile(fmt.Sprintf(`{
								"adapter":  "postgres",
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
						fakePgClient.WhenCalled().WillPrintToStdOut(
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

							Expect(fakePgClient.Invocations()).To(HaveLen(2))
							Expect(fakePgClient.Invocations()[1].Args()).Should(ConsistOf(expectedArgs))
							Expect(fakePgClient.Invocations()[1].Env()).Should(HaveKeyWithValue("PGPASSWORD", password))
						})

						By("exiting with a helpful error message", func() {
							Expect(session).Should(gexec.Exit(1))
							Expect(session.Err).Should(gbytes.Say(`can't find specified table\(s\): table3`))
						})
					})
				})

			})

			Context("when pg_dump fails", func() {
				BeforeEach(func() {
					configFile = saveFile(fmt.Sprintf(`{
								"adapter":  "postgres",
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
					fakePgClient.WhenCalled().WillPrintToStdOut(
						" table1 \n table2 \n\n\n").
						WillExitWith(0)
					fakePgDump10.WhenCalled().WillExitWith(1)
				})

				It("also fails", func() {
					Eventually(session).Should(gexec.Exit(1))
				})
			})

		})

		Context("Postgres database server is version 11", func() {
			BeforeEach(func() {
				fakePgClient.WhenCalled().WillPrintToStdOut(
					" PostgreSQL 11.2 on x86_64-pc-linux-gnu, compiled by gcc " +
						"(Ubuntu 5.4.0-6ubuntu1~16.04.11) 5.4.0 20160609, 64-bit").
					WillExitWith(0)
			})

			Context("when pg_dump succeeds", func() {
				BeforeEach(func() {
					fakePgDump11.WhenCalled().WillExitWith(0)
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

						Expect(fakePgClient.Invocations()).To(HaveLen(1))
						Expect(fakePgClient.Invocations()[0].Args()).Should(ConsistOf(expectedArgs))
						Expect(fakePgClient.Invocations()[0].Env()).Should(HaveKeyWithValue("PGPASSWORD", password))
					})

					By("dumping the database with the correct dump binary", func() {
						expectedArgs := []string{
							"--verbose",
							fmt.Sprintf("--username=%s", username),
							fmt.Sprintf("--host=%s", host),
							fmt.Sprintf("--port=%d", port),
							"--format=custom",
							fmt.Sprintf("--file=%s", artifactFile),
							databaseName,
						}

						Expect(fakePgDump11.Invocations()).To(HaveLen(1))
						Expect(fakePgDump11.Invocations()[0].Args()).Should(ConsistOf(expectedArgs))
						Expect(fakePgDump11.Invocations()[0].Env()).Should(HaveKeyWithValue("PGPASSWORD", password))
					})

					By("not invoking the dump binary for a different version", func() {
						Expect(fakePgDump15.Invocations()).To(HaveLen(0))
					})

					Expect(session).Should(gexec.Exit(0))
				})

				Context("when 'tables' are specified in the configFile", func() {
					BeforeEach(func() {
						configFile = saveFile(fmt.Sprintf(`{
								"adapter":  "postgres",
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
						fakePgClient.WhenCalled().WillPrintToStdOut(
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

							Expect(fakePgClient.Invocations()).To(HaveLen(2))
							Expect(fakePgClient.Invocations()[1].Args()).Should(ConsistOf(expectedArgs))
							Expect(fakePgClient.Invocations()[1].Env()).Should(HaveKeyWithValue("PGPASSWORD", password))
						})

						By("calling pg_dump with the correct arguments", func() {
							expectedArgs := []string{
								"--verbose",
								fmt.Sprintf("--username=%s", username),
								fmt.Sprintf("--host=%s", host),
								fmt.Sprintf("--port=%d", port),
								"--format=custom",
								fmt.Sprintf("--file=%s", artifactFile),
								databaseName,
								"-t", "table1",
								"-t", "table2",
								"-t", "table3",
							}

							Expect(fakePgDump11.Invocations()[0].Args()).Should(ConsistOf(expectedArgs))
						})

						Expect(session).Should(gexec.Exit(0))
					})
				})

				Context("when missing 'tables' are specified in the configFile", func() {
					BeforeEach(func() {
						configFile = saveFile(fmt.Sprintf(`{
								"adapter":  "postgres",
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
						fakePgClient.WhenCalled().WillPrintToStdOut(
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

							Expect(fakePgClient.Invocations()).To(HaveLen(2))
							Expect(fakePgClient.Invocations()[1].Args()).Should(ConsistOf(expectedArgs))
							Expect(fakePgClient.Invocations()[1].Env()).Should(HaveKeyWithValue("PGPASSWORD", password))
						})

						By("exiting with a helpful error message", func() {
							Expect(session).Should(gexec.Exit(1))
							Expect(session.Err).Should(gbytes.Say(`can't find specified table\(s\): table3`))
						})
					})
				})

			})

			Context("when pg_dump fails", func() {
				BeforeEach(func() {
					configFile = saveFile(fmt.Sprintf(`{
								"adapter":  "postgres",
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
					fakePgClient.WhenCalled().WillPrintToStdOut(
						" table1 \n table2 \n\n\n").
						WillExitWith(0)
					fakePgDump11.WhenCalled().WillExitWith(1)
				})

				It("also fails", func() {
					Eventually(session).Should(gexec.Exit(1))
				})
			})

		})

		Context("Postgres database server is version 13", func() {
			BeforeEach(func() {
				fakePgClient.WhenCalled().WillPrintToStdOut(
					" PostgreSQL 13.2 on x86_64-pc-linux-gnu, compiled by gcc " +
						"(Ubuntu 5.4.0-6ubuntu1~16.04.12) 5.4.0 20160609, 64-bit").
					WillExitWith(0)
			})

			Context("when pg_dump succeeds", func() {
				BeforeEach(func() {
					fakePgDump13.WhenCalled().WillExitWith(0)
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

						Expect(fakePgClient.Invocations()).To(HaveLen(1))
						Expect(fakePgClient.Invocations()[0].Args()).Should(ConsistOf(expectedArgs))
						Expect(fakePgClient.Invocations()[0].Env()).Should(HaveKeyWithValue("PGPASSWORD", password))
					})

					By("dumping the database with the correct dump binary", func() {
						expectedArgs := []string{
							"--verbose",
							fmt.Sprintf("--username=%s", username),
							fmt.Sprintf("--host=%s", host),
							fmt.Sprintf("--port=%d", port),
							"--format=custom",
							fmt.Sprintf("--file=%s", artifactFile),
							databaseName,
						}

						Expect(fakePgDump13.Invocations()).To(HaveLen(1))
						Expect(fakePgDump13.Invocations()[0].Args()).Should(ConsistOf(expectedArgs))
						Expect(fakePgDump13.Invocations()[0].Env()).Should(HaveKeyWithValue("PGPASSWORD", password))
					})

					By("not invoking the dump binary for a different version", func() {
						Expect(fakePgDump11.Invocations()).To(HaveLen(0))
					})

					Expect(session).Should(gexec.Exit(0))
				})

				Context("when 'tables' are specified in the configFile", func() {
					BeforeEach(func() {
						configFile = saveFile(fmt.Sprintf(`{
								"adapter":  "postgres",
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
						fakePgClient.WhenCalled().WillPrintToStdOut(
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

							Expect(fakePgClient.Invocations()).To(HaveLen(2))
							Expect(fakePgClient.Invocations()[1].Args()).Should(ConsistOf(expectedArgs))
							Expect(fakePgClient.Invocations()[1].Env()).Should(HaveKeyWithValue("PGPASSWORD", password))
						})

						By("calling pg_dump with the correct arguments", func() {
							expectedArgs := []string{
								"--verbose",
								fmt.Sprintf("--username=%s", username),
								fmt.Sprintf("--host=%s", host),
								fmt.Sprintf("--port=%d", port),
								"--format=custom",
								fmt.Sprintf("--file=%s", artifactFile),
								databaseName,
								"-t", "table1",
								"-t", "table2",
								"-t", "table3",
							}

							Expect(fakePgDump13.Invocations()[0].Args()).Should(ConsistOf(expectedArgs))
						})

						Expect(session).Should(gexec.Exit(0))
					})
				})

				Context("when missing 'tables' are specified in the configFile", func() {
					BeforeEach(func() {
						configFile = saveFile(fmt.Sprintf(`{
								"adapter":  "postgres",
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
						fakePgClient.WhenCalled().WillPrintToStdOut(
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

							Expect(fakePgClient.Invocations()).To(HaveLen(2))
							Expect(fakePgClient.Invocations()[1].Args()).Should(ConsistOf(expectedArgs))
							Expect(fakePgClient.Invocations()[1].Env()).Should(HaveKeyWithValue("PGPASSWORD", password))
						})

						By("exiting with a helpful error message", func() {
							Expect(session).Should(gexec.Exit(1))
							Expect(session.Err).Should(gbytes.Say(`can't find specified table\(s\): table3`))
						})
					})
				})

			})

			Context("when pg_dump fails", func() {
				BeforeEach(func() {
					configFile = saveFile(fmt.Sprintf(`{
								"adapter":  "postgres",
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
					fakePgClient.WhenCalled().WillPrintToStdOut(
						" table1 \n table2 \n\n\n").
						WillExitWith(0)
					fakePgDump13.WhenCalled().WillExitWith(1)
				})

				It("also fails", func() {
					Eventually(session).Should(gexec.Exit(1))
				})
			})

		})

		Context("Postgres database server is version 15", func() {
			BeforeEach(func() {
				fakePgClient.WhenCalled().WillPrintToStdOut(
					" PostgreSQL 15.3 on x86_64-pc-linux-gnu, compiled by gcc " +
						"(Ubuntu 5.4.0-6ubuntu1~16.04.12) 5.4.0 20160609, 64-bit").
					WillExitWith(0)
			})

			Context("when pg_dump succeeds", func() {
				BeforeEach(func() {
					fakePgDump15.WhenCalled().WillExitWith(0)
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

						Expect(fakePgClient.Invocations()).To(HaveLen(1))
						Expect(fakePgClient.Invocations()[0].Args()).Should(ConsistOf(expectedArgs))
						Expect(fakePgClient.Invocations()[0].Env()).Should(HaveKeyWithValue("PGPASSWORD", password))
					})

					By("dumping the database with the correct dump binary", func() {
						expectedArgs := []string{
							"--verbose",
							fmt.Sprintf("--username=%s", username),
							fmt.Sprintf("--host=%s", host),
							fmt.Sprintf("--port=%d", port),
							"--format=custom",
							fmt.Sprintf("--file=%s", artifactFile),
							databaseName,
						}

						Expect(fakePgDump15.Invocations()).To(HaveLen(1))
						Expect(fakePgDump15.Invocations()[0].Args()).Should(ConsistOf(expectedArgs))
						Expect(fakePgDump15.Invocations()[0].Env()).Should(HaveKeyWithValue("PGPASSWORD", password))
					})

					By("not invoking the dump binary for a different version", func() {
						Expect(fakePgDump11.Invocations()).To(HaveLen(0))
					})

					Expect(session).Should(gexec.Exit(0))
				})

				Context("when 'tables' are specified in the configFile", func() {
					BeforeEach(func() {
						configFile = saveFile(fmt.Sprintf(`{
								"adapter":  "postgres",
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
						fakePgClient.WhenCalled().WillPrintToStdOut(
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

							Expect(fakePgClient.Invocations()).To(HaveLen(2))
							Expect(fakePgClient.Invocations()[1].Args()).Should(ConsistOf(expectedArgs))
							Expect(fakePgClient.Invocations()[1].Env()).Should(HaveKeyWithValue("PGPASSWORD", password))
						})

						By("calling pg_dump with the correct arguments", func() {
							expectedArgs := []string{
								"--verbose",
								fmt.Sprintf("--username=%s", username),
								fmt.Sprintf("--host=%s", host),
								fmt.Sprintf("--port=%d", port),
								"--format=custom",
								fmt.Sprintf("--file=%s", artifactFile),
								databaseName,
								"-t", "table1",
								"-t", "table2",
								"-t", "table3",
							}

							Expect(fakePgDump15.Invocations()[0].Args()).Should(ConsistOf(expectedArgs))
						})

						Expect(session).Should(gexec.Exit(0))
					})
				})

				Context("when missing 'tables' are specified in the configFile", func() {
					BeforeEach(func() {
						configFile = saveFile(fmt.Sprintf(`{
								"adapter":  "postgres",
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
						fakePgClient.WhenCalled().WillPrintToStdOut(
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

							Expect(fakePgClient.Invocations()).To(HaveLen(2))
							Expect(fakePgClient.Invocations()[1].Args()).Should(ConsistOf(expectedArgs))
							Expect(fakePgClient.Invocations()[1].Env()).Should(HaveKeyWithValue("PGPASSWORD", password))
						})

						By("exiting with a helpful error message", func() {
							Expect(session).Should(gexec.Exit(1))
							Expect(session.Err).Should(gbytes.Say(`can't find specified table\(s\): table3`))
						})
					})
				})

			})

			Context("when pg_dump fails", func() {
				BeforeEach(func() {
					configFile = saveFile(fmt.Sprintf(`{
								"adapter":  "postgres",
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
					fakePgClient.WhenCalled().WillPrintToStdOut(
						" table1 \n table2 \n\n\n").
						WillExitWith(0)
					fakePgDump15.WhenCalled().WillExitWith(1)
				})

				It("also fails", func() {
					Eventually(session).Should(gexec.Exit(1))
				})
			})

		})
	})

	Context("restore", func() {
		JustBeforeEach(func() {
			session = run(compiledSDKPath, envVars,
				"--artifact-file", artifactFile,
				"--config", configFile.Name(),
				"--restore",
			)
		})

		Context("Postgres database server is version 10", func() {
			BeforeEach(func() {
				fakePgClient.WhenCalled().WillPrintToStdOut(
					" PostgreSQL 10.6 on x86_64-pc-linux-gnu, compiled by gcc " +
						"(Ubuntu 5.4.0-6ubuntu1~16.04.11) 5.4.0 20160609, 64-bit").
					WillExitWith(0)
			})

			Context("pg_restore succeeds", func() {
				BeforeEach(func() {
					fakePgRestore10.WhenCalled().WillExitWith(0)
					fakePgRestore10.WhenCalled().WillExitWith(0)
				})

				It("calls pg_restore to get information about the restore", func() {
					Expect(fakePgRestore10.Invocations()).To(HaveLen(2))

					Expect(fakePgRestore10.Invocations()[0].Args()).To(Equal([]string{"--list", artifactFile}))

					expectedArgs := []interface{}{
						"--verbose",
						fmt.Sprintf("--username=%s", username),
						fmt.Sprintf("--host=%s", host),
						fmt.Sprintf("--port=%d", port),
						"--format=custom",
						fmt.Sprintf("--dbname=%s", databaseName),
						"--clean",
						"--if-exists",
						"--single-transaction",
						"--exit-on-error",
						HavePrefix("--use-list="),
						artifactFile,
					}

					Expect(fakePgRestore10.Invocations()[1].Args()).Should(ConsistOf(expectedArgs))
					Expect(fakePgRestore10.Invocations()[1].Env()).Should(HaveKeyWithValue("PGPASSWORD", password))
				})

				It("succeeds", func() {
					Expect(session).Should(gexec.Exit(0))
				})
			})

			Context("and pg_restore fails when restoring", func() {
				BeforeEach(func() {
					fakePgRestore10.WhenCalled().WillExitWith(0)
					fakePgRestore10.WhenCalled().WillExitWith(1)
				})

				It("also fails", func() {
					Eventually(session).Should(gexec.Exit(1))
				})
			})

			Context("and pg_restore fails to get file list", func() {
				BeforeEach(func() {
					fakePgRestore10.WhenCalled().WillExitWith(1)
				})

				It("also fails", func() {
					Eventually(session).Should(gexec.Exit(1))
				})
			})
		})

		Context("Postgres database server is version 11", func() {
			BeforeEach(func() {
				fakePgClient.WhenCalled().WillPrintToStdOut(
					" PostgreSQL 11.2 on x86_64-pc-linux-gnu, compiled by gcc " +
						"(Ubuntu 5.4.0-6ubuntu1~16.04.11) 5.4.0 20160609, 64-bit").
					WillExitWith(0)
			})

			Context("pg_restore succeeds", func() {
				BeforeEach(func() {
					fakePgRestore11.WhenCalled().WillExitWith(0)
					fakePgRestore11.WhenCalled().WillExitWith(0)
				})

				It("calls pg_restore to get information about the restore", func() {
					Expect(fakePgRestore11.Invocations()).To(HaveLen(2))

					Expect(fakePgRestore11.Invocations()[0].Args()).To(Equal([]string{"--list", artifactFile}))

					expectedArgs := []interface{}{
						"--verbose",
						fmt.Sprintf("--username=%s", username),
						fmt.Sprintf("--host=%s", host),
						fmt.Sprintf("--port=%d", port),
						"--format=custom",
						fmt.Sprintf("--dbname=%s", databaseName),
						"--clean",
						"--if-exists",
						"--single-transaction",
						"--exit-on-error",
						HavePrefix("--use-list="),
						artifactFile,
					}

					Expect(fakePgRestore11.Invocations()[1].Args()).Should(ConsistOf(expectedArgs))
					Expect(fakePgRestore11.Invocations()[1].Env()).Should(HaveKeyWithValue("PGPASSWORD", password))
				})

				It("succeeds", func() {
					Expect(session).Should(gexec.Exit(0))
				})
			})

			Context("and pg_restore fails when restoring", func() {
				BeforeEach(func() {
					fakePgRestore11.WhenCalled().WillExitWith(0)
					fakePgRestore11.WhenCalled().WillExitWith(1)
				})

				It("also fails", func() {
					Eventually(session).Should(gexec.Exit(1))
				})
			})

			Context("and pg_restore fails to get file list", func() {
				BeforeEach(func() {
					fakePgRestore11.WhenCalled().WillExitWith(1)
				})

				It("also fails", func() {
					Eventually(session).Should(gexec.Exit(1))
				})
			})
		})

		Context("Postgres database server is version 13", func() {
			BeforeEach(func() {
				fakePgClient.WhenCalled().WillPrintToStdOut(
					" PostgreSQL 13.2 on x86_64-pc-linux-gnu, compiled by gcc " +
						"(Ubuntu 5.4.0-6ubuntu1~16.04.12) 5.4.0 20160609, 64-bit").
					WillExitWith(0)
			})

			Context("pg_restore succeeds", func() {
				BeforeEach(func() {
					fakePgRestore13.WhenCalled().WillExitWith(0)
					fakePgRestore13.WhenCalled().WillExitWith(0)
				})

				It("calls pg_restore to get information about the restore", func() {
					Expect(fakePgRestore13.Invocations()).To(HaveLen(2))

					Expect(fakePgRestore13.Invocations()[0].Args()).To(Equal([]string{"--list", artifactFile}))

					expectedArgs := []interface{}{
						"--verbose",
						fmt.Sprintf("--username=%s", username),
						fmt.Sprintf("--host=%s", host),
						fmt.Sprintf("--port=%d", port),
						"--format=custom",
						fmt.Sprintf("--dbname=%s", databaseName),
						"--clean",
						"--if-exists",
						"--single-transaction",
						"--exit-on-error",
						HavePrefix("--use-list="),
						artifactFile,
					}

					Expect(fakePgRestore13.Invocations()[1].Args()).Should(ConsistOf(expectedArgs))
					Expect(fakePgRestore13.Invocations()[1].Env()).Should(HaveKeyWithValue("PGPASSWORD", password))
				})

				It("succeeds", func() {
					Expect(session).Should(gexec.Exit(0))
				})
			})

			Context("and pg_restore fails when restoring", func() {
				BeforeEach(func() {
					fakePgRestore13.WhenCalled().WillExitWith(0)
					fakePgRestore13.WhenCalled().WillExitWith(1)
				})

				It("also fails", func() {
					Eventually(session).Should(gexec.Exit(1))
				})
			})

			Context("and pg_restore fails to get file list", func() {
				BeforeEach(func() {
					fakePgRestore13.WhenCalled().WillExitWith(1)
				})

				It("also fails", func() {
					Eventually(session).Should(gexec.Exit(1))
				})
			})
		})

		Context("Postgres database server is version 15", func() {
			BeforeEach(func() {
				fakePgClient.WhenCalled().WillPrintToStdOut(
					" PostgreSQL 15.3 on x86_64-pc-linux-gnu, compiled by gcc " +
						"(Ubuntu 5.4.0-6ubuntu1~16.04.12) 5.4.0 20160609, 64-bit").
					WillExitWith(0)
			})

			Context("pg_restore succeeds", func() {
				BeforeEach(func() {
					fakePgRestore15.WhenCalled().WillExitWith(0)
					fakePgRestore15.WhenCalled().WillExitWith(0)
				})

				It("calls pg_restore to get information about the restore", func() {
					Expect(fakePgRestore15.Invocations()).To(HaveLen(2))

					Expect(fakePgRestore15.Invocations()[0].Args()).To(Equal([]string{"--list", artifactFile}))

					expectedArgs := []interface{}{
						"--verbose",
						fmt.Sprintf("--username=%s", username),
						fmt.Sprintf("--host=%s", host),
						fmt.Sprintf("--port=%d", port),
						"--format=custom",
						fmt.Sprintf("--dbname=%s", databaseName),
						"--clean",
						"--if-exists",
						"--single-transaction",
						"--exit-on-error",
						HavePrefix("--use-list="),
						artifactFile,
					}

					Expect(fakePgRestore15.Invocations()[1].Args()).Should(ConsistOf(expectedArgs))
					Expect(fakePgRestore15.Invocations()[1].Env()).Should(HaveKeyWithValue("PGPASSWORD", password))
				})

				It("succeeds", func() {
					Expect(session).Should(gexec.Exit(0))
				})
			})

			Context("and pg_restore fails when restoring", func() {
				BeforeEach(func() {
					fakePgRestore15.WhenCalled().WillExitWith(0)
					fakePgRestore15.WhenCalled().WillExitWith(1)
				})

				It("also fails", func() {
					Eventually(session).Should(gexec.Exit(1))
				})
			})

			Context("and pg_restore fails to get file list", func() {
				BeforeEach(func() {
					fakePgRestore15.WhenCalled().WillExitWith(1)
				})

				It("also fails", func() {
					Eventually(session).Should(gexec.Exit(1))
				})
			})
		})
	})
})

func run(path string, env map[string]string, args ...string) *gexec.Session {
	cmd := exec.Command(path, args...)
	for key, val := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, val))
	}
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).ToNot(HaveOccurred())
	Eventually(session).Should(gexec.Exit())

	return session
}
