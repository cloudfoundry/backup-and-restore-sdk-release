package postgresql

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"strings"

	. "github.com/cloudfoundry-incubator/database-backup-restore/system_tests/utils"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("postgres with tls", func() {
	if os.Getenv("TEST_TLS") == "false" {
		fmt.Println("**********************************************")
		fmt.Println("Not testing TLS")
		fmt.Println("**********************************************")
		return
	}
	var dbDumpPath string
	var configPath string
	var databaseName string
	var configJson string

	BeforeEach(func() {
		disambiguationString := DisambiguationString()
		configPath = "/tmp/config" + disambiguationString
		dbDumpPath = "/tmp/artifact" + disambiguationString
		databaseName = "db" + disambiguationString

		pgConnection = NewPostgresConnection(
			postgresHostName,
			postgresPort,
			postgresNonSslUsername,
			postgresPassword,
			os.Getenv("SSH_PROXY_HOST"),
			os.Getenv("SSH_PROXY_USER"),
			os.Getenv("SSH_PROXY_KEY_FILE"),
		)

		pgConnection.Open("postgres")
		pgConnection.RunSQLCommand("CREATE DATABASE " + databaseName)
		pgConnection.SwitchToDb(databaseName)
		pgConnection.RunSQLCommand("CREATE TABLE people (name varchar(255));")
		pgConnection.RunSQLCommand("INSERT INTO people VALUES ('Old Person');")
	})

	AfterEach(func() {
		pgConnection.SwitchToDb("postgres")
		pgConnection.RunSQLCommand("DROP DATABASE " + databaseName)
		pgConnection.Close()

		brJob.RunOnVMAndSucceed(fmt.Sprintf("sudo rm -rf %s %s", configPath, dbDumpPath))
	})

	JustBeforeEach(func() {
		brJob.RunOnVMAndSucceed(fmt.Sprintf("echo '%s' > %s", configJson, configPath))
	})

	Context("when the db user requires TLS", func() {
		BeforeEach(func() {
			if os.Getenv("TEST_SSL_USER_REQUIRES_SSL") == "false" {
				return
			}

			_, _, err := ConnectToPostgresWithNoSsl(
				postgresHostName,
				postgresPassword,
				postgresSslUsername,
				postgresPort,
				databaseName,
				os.Getenv("SSH_PROXY_HOST"),
				os.Getenv("SSH_PROXY_USER"),
				os.Getenv("SSH_PROXY_KEY_FILE"),
			)

			Expect(err).To(MatchError(MatchRegexp("no pg_hba.conf entry for host \".*\", user \"ssl_user\", database \".*\", SSL off")))
		})

		Context("when TLS info is not provided in the config", func() {
			BeforeEach(func() {
				configJson = fmt.Sprintf(
					`{
						"username": "%s",
						"password": "%s",
						"host": "%s",
						"port": %d,
						"database": "%s",
						"adapter": "postgres"
					}`,
					postgresSslUsername,
					postgresPassword,
					postgresHostName,
					postgresPort,
					databaseName,
				)
			})

			It("works", func() {
				brJob.RunOnVMAndSucceed(
					fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/backup --artifact-file %s --config %s",
						dbDumpPath, configPath))

				pgConnection.RunSQLCommand("UPDATE people SET NAME = 'New Person';")

				brJob.RunOnVMAndSucceed(
					fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/restore --artifact-file %s --config %s",
						dbDumpPath, configPath))

				Expect(pgConnection.FetchSQLColumn("SELECT name FROM people;")).To(ConsistOf("Old Person"))
			})
		})

		Context("when TLS info is provided in the config", func() {
			Context("And host verification is not skipped", func() {
				if os.Getenv("TEST_TLS_VERIFY_IDENTITY") == "false" {
					fmt.Println("**********************************************")
					fmt.Println("Not testing TLS with Verify Identity")
					fmt.Println("**********************************************")
					return
				}

				Context("And the CA cert is correct", func() {
					BeforeEach(func() {
						configJson = fmt.Sprintf(
							`{
							"username": "%s",
							"password": "%s",
							"host": "%s",
							"port": %d,
							"database": "%s",
							"adapter": "postgres",
							"tls": {
								"cert": {
									"ca": "%s"
								}
							}
						}`,
							postgresSslUsername,
							postgresPassword,
							postgresHostName,
							postgresPort,
							databaseName,
							escapeNewLines(postgresCaCert),
						)
					})

					It("works", func() {
						brJob.RunOnVMAndSucceed(
							fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/backup --artifact-file %s --config %s",
								dbDumpPath, configPath))

						pgConnection.RunSQLCommand("UPDATE people SET NAME = 'New Person';")

						brJob.RunOnVMAndSucceed(
							fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/restore --artifact-file %s --config %s",
								dbDumpPath, configPath))

						Expect(pgConnection.FetchSQLColumn("SELECT name FROM people;")).To(ConsistOf("Old Person"))
					})
				})
			})

			Context("And host verification is skipped", func() {
				Context("And the CA cert is correct", func() {
					BeforeEach(func() {
						configJson = fmt.Sprintf(
							`{
							"username": "%s",
							"password": "%s",
							"host": "%s",
							"port": %d,
							"database": "%s",
							"adapter": "postgres",
							"tls": {
								"skip_host_verify": true,
								"cert": {
									"ca": "%s"
								}
							}
						}`,
							postgresSslUsername,
							postgresPassword,
							postgresHostName,
							postgresPort,
							databaseName,
							escapeNewLines(postgresCaCert),
						)
					})

					It("works", func() {
						brJob.RunOnVMAndSucceed(
							fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/backup --artifact-file %s --config %s",
								dbDumpPath, configPath))

						pgConnection.RunSQLCommand("UPDATE people SET NAME = 'New Person';")

						brJob.RunOnVMAndSucceed(
							fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/restore --artifact-file %s --config %s",
								dbDumpPath, configPath))

						Expect(pgConnection.FetchSQLColumn("SELECT name FROM people;")).To(ConsistOf("Old Person"))
					})
				})

				Context("And the CA cert is malformed", func() {
					BeforeEach(func() {
						configJson = fmt.Sprintf(
							`{
									"username": "%s",
									"password": "%s",
									"host": "%s",
									"port": %d,
									"database": "%s",
									"adapter": "postgres",
									"tls": {
										"skip_host_verify": true,
										"cert": {
											"ca": "fooooooo"
										}
									}
								}`,
							postgresSslUsername,
							postgresPassword,
							postgresHostName,
							postgresPort,
							databaseName,
						)
					})

					It("does not work", func() {
						Expect(brJob.RunOnInstance("/var/vcap/jobs/database-backup-restorer/bin/backup",
							"--artifact-file", dbDumpPath, "--config", configPath)).To(gexec.Exit(1))
					})
				})
			})
		})
	})

	Context("when the db user requires TLS and Mutual TLS", func() {
		if os.Getenv("TEST_TLS_MUTUAL_TLS") == "false" {
			fmt.Println("**********************************************")
			fmt.Println("Not testing TLS with Mutual TLS")
			fmt.Println("**********************************************")
			return
		}

		Context("when TLS info is not provided in the config", func() {
			BeforeEach(func() {
				configJson = fmt.Sprintf(
					`{
						"username": "%s",
						"password": "%s",
						"host": "%s",
						"port": %d,
						"database": "%s",
						"adapter": "postgres"
					}`,
					postgresMutualTlsUsername,
					postgresPassword,
					postgresHostName,
					postgresPort,
					databaseName,
				)
			})

			It("does not work", func() {
				Expect(brJob.RunOnInstance("/var/vcap/jobs/database-backup-restorer/bin/backup",
					"--artifact-file", dbDumpPath, "--config", configPath)).To(gexec.Exit(1))
			})
		})

		Context("when TLS info is provided in the config", func() {
			Context("and host verification is not skipped", func() {
				if os.Getenv("TEST_TLS_VERIFY_IDENTITY") == "false" {
					return
				}

				Context("and the CA cert is correct", func() {
					BeforeEach(func() {
						configJson = fmt.Sprintf(
							`{
							"username": "%s",
							"password": "%s",
							"host": "%s",
							"port": %d,
							"database": "%s",
							"adapter": "postgres",
							"tls": {
								"cert": {
									"ca": "%s",
									"certificate": "%s",
									"private_key": "%s"
								}
							}
						}`,
							postgresMutualTlsUsername,
							postgresPassword,
							postgresHostName,
							postgresPort,
							databaseName,
							escapeNewLines(postgresCaCert),
							escapeNewLines(postgresClientCert),
							escapeNewLines(postgresClientKey),
						)
					})

					It("works", func() {
						brJob.RunOnVMAndSucceed(
							fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/backup --artifact-file %s --config %s",
								dbDumpPath, configPath))

						pgConnection.RunSQLCommand("UPDATE people SET NAME = 'New Person';")

						brJob.RunOnVMAndSucceed(
							fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/restore --artifact-file %s --config %s",
								dbDumpPath, configPath))

						Expect(pgConnection.FetchSQLColumn("SELECT name FROM people;")).To(ConsistOf("Old Person"))
					})
				})
			})

			Context("and host verification is skipped", func() {
				Context("and the client cert and key are provided and correct", func() {
					BeforeEach(func() {
						configJson = fmt.Sprintf(
							`{
									"username": "%s",
									"password": "%s",
									"host": "%s",
									"port": %d,
									"database": "%s",
									"adapter": "postgres",
									"tls": {
										"skip_host_verify": true,
										"cert": {
											"ca": "%s",
											"certificate": "%s",
											"private_key": "%s"
										}
									}
								}`,
							postgresMutualTlsUsername,
							postgresPassword,
							postgresHostName,
							postgresPort,
							databaseName,
							escapeNewLines(postgresCaCert),
							escapeNewLines(postgresClientCert),
							escapeNewLines(postgresClientKey),
						)
					})

					It("works", func() {
						brJob.RunOnVMAndSucceed(
							fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/backup --artifact-file %s --config %s",
								dbDumpPath, configPath))

						pgConnection.RunSQLCommand("UPDATE people SET NAME = 'New Person';")

						brJob.RunOnVMAndSucceed(
							fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/restore --artifact-file %s --config %s",
								dbDumpPath, configPath))

						Expect(pgConnection.FetchSQLColumn("SELECT name FROM people;")).To(ConsistOf("Old Person"))
					})
				})

				Context("and the client cert and key are provided and malformed", func() {
					BeforeEach(func() {
						configJson = fmt.Sprintf(
							`{
									"username": "%s",
									"password": "%s",
									"host": "%s",
									"port": %d,
									"database": "%s",
									"adapter": "postgres",
									"tls": {
										"skip_host_verify": true,
										"cert": {
											"ca": "%s",
											"certificate": "foo",
											"private_key": "bar"
										}
									}
								}`,
							postgresMutualTlsUsername,
							postgresPassword,
							postgresHostName,
							postgresPort,
							databaseName,
							escapeNewLines(postgresCaCert),
						)
					})

					It("does not work", func() {
						Expect(brJob.RunOnInstance("/var/vcap/jobs/database-backup-restorer/bin/backup",
							"--artifact-file", dbDumpPath, "--config", configPath)).To(gexec.Exit(1))
					})
				})

				Context("and the client cert and key are not provided", func() {
					BeforeEach(func() {
						configJson = fmt.Sprintf(
							`{
							"username": "%s",
							"password": "%s",
							"host": "%s",
							"port": %d,
							"database": "%s",
							"adapter": "postgres",
							"tls": {
								"skip_host_verify": true,
								"cert": {
									"ca": "%s"
								}
							}
						}`,
							postgresMutualTlsUsername,
							postgresPassword,
							postgresHostName,
							postgresPort,
							databaseName,
							escapeNewLines(postgresCaCert),
						)
					})

					It("does not work", func() {
						Expect(brJob.RunOnInstance("/var/vcap/jobs/database-backup-restorer/bin/backup",
							"--artifact-file", dbDumpPath, "--config", configPath)).To(gexec.Exit(1))
					})
				})
			})
		})
	})

	Context("when the db user does not require TLS", func() {
		Context("and host verification is not skipped", func() {
			if os.Getenv("TEST_TLS_VERIFY_IDENTITY") == "false" {
				return
			}

			Context("and the correct CA cert is provided", func() {
				BeforeEach(func() {
					configJson = fmt.Sprintf(
						`{
						"username": "%s",
						"password": "%s",
						"host": "%s",
						"port": %d,
						"database": "%s",
						"adapter": "postgres",
						"tls": {
							"cert": {
								"ca": "%s"
							}
						}
					}`,
						postgresNonSslUsername,
						postgresPassword,
						postgresHostName,
						postgresPort,
						databaseName,
						escapeNewLines(postgresCaCert),
					)
				})

				It("works", func() {
					brJob.RunOnVMAndSucceed(
						fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/backup --artifact-file %s --config %s",
							dbDumpPath, configPath))

					pgConnection.RunSQLCommand("UPDATE people SET NAME = 'New Person';")

					brJob.RunOnVMAndSucceed(
						fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/restore --artifact-file %s --config %s",
							dbDumpPath, configPath))

					Expect(pgConnection.FetchSQLColumn("SELECT name FROM people;")).To(ConsistOf("Old Person"))
				})
			})
		})

		Context("and host verification is skipped", func() {
			Context("and the correct CA cert is provided", func() {
				BeforeEach(func() {
					configJson = fmt.Sprintf(
						`{
						"username": "%s",
						"password": "%s",
						"host": "%s",
						"port": %d,
						"database": "%s",
						"adapter": "postgres",
						"tls": {
							"skip_host_verify": true,
							"cert": {
								"ca": "%s"
							}
						}
					}`,
						postgresNonSslUsername,
						postgresPassword,
						postgresHostName,
						postgresPort,
						databaseName,
						escapeNewLines(postgresCaCert),
					)
				})

				It("works", func() {
					brJob.RunOnVMAndSucceed(
						fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/backup --artifact-file %s --config %s",
							dbDumpPath, configPath))

					pgConnection.RunSQLCommand("UPDATE people SET NAME = 'New Person';")

					brJob.RunOnVMAndSucceed(
						fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/restore --artifact-file %s --config %s",
							dbDumpPath, configPath))

					Expect(pgConnection.FetchSQLColumn("SELECT name FROM people;")).To(ConsistOf("Old Person"))
				})
			})

			Context("and the CA cert is malformed", func() {
				BeforeEach(func() {
					configJson = fmt.Sprintf(
						`{
								"username": "%s",
								"password": "%s",
								"host": "%s",
								"port": %d,
								"database": "%s",
								"adapter": "postgres",
								"tls": {
									"skip_host_verify": true,
									"cert": {
										"ca": "fooooooo"
									}
								}
							}`,
						postgresNonSslUsername,
						postgresPassword,
						postgresHostName,
						postgresPort,
						databaseName,
					)
				})

				It("does not work", func() {
					Expect(brJob.RunOnInstance("/var/vcap/jobs/database-backup-restorer/bin/backup",
						"--artifact-file", dbDumpPath, "--config", configPath)).To(gexec.Exit(1))
				})
			})
		})
	})
})

func escapeNewLines(txt string) string {
	return strings.Replace(txt, "\n", "\\n", -1)
}
