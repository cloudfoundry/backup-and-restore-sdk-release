package mysql

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	_ "github.com/go-sql-driver/mysql"

	"strings"

	. "github.com/cloudfoundry-incubator/database-backup-restore/system_tests/utils"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("mysql with tls", func() {
	if os.Getenv("TEST_TLS") == "false" {
		fmt.Println("**********************************************")
		fmt.Println("Not testing TLS")
		fmt.Println("**********************************************")
		return
	}
	var dbDumpPath string
	var configPath string
	var databaseName string
	var mysqlSslUsername string
	var configJson string

	BeforeEach(func() {
		disambiguationString := DisambiguationString()
		configPath = "/tmp/config" + disambiguationString
		dbDumpPath = "/tmp/artifact" + disambiguationString
		databaseName = "db" + disambiguationString

		RunSQLCommand("CREATE DATABASE "+databaseName, connection)
		RunSQLCommand("USE "+databaseName, connection)
		RunSQLCommand("CREATE TABLE people (name varchar(255));", connection)
		RunSQLCommand("INSERT INTO people VALUES ('Old Person');", connection)

		mysqlSslUsername = "ssl_user_" + DisambiguationStringOfLength(6)
		RunSQLCommand(fmt.Sprintf(
			"CREATE USER '%s' IDENTIFIED BY '%s';",
			mysqlSslUsername, mysqlPassword), connection)
	})

	AfterEach(func() {
		RunSQLCommand(fmt.Sprintf("DROP USER '%s';", mysqlSslUsername), connection)
	})

	JustBeforeEach(func() {
		brJob.RunOnVMAndSucceed(fmt.Sprintf("echo '%s' > %s", configJson, configPath))
	})

	Context("when the db user requires TLS", func() {
		BeforeEach(func() {
			RunSQLCommand(fmt.Sprintf("GRANT ALL PRIVILEGES ON %s.* TO %s REQUIRE SSL;", databaseName, mysqlSslUsername), connection)
		})

		Context("when TLS info is not provided in the config", func() {
			BeforeEach(func() {
				configJson = fmt.Sprintf(
					`{
						"username": "%s",
						"password": "%s",
						"host": "%s",
						"port": %s,
						"database": "%s",
						"adapter": "mysql"
					}`,
					mysqlSslUsername,
					mysqlPassword,
					mysqlHostName,
					mysqlPort,
					databaseName,
				)
			})

			It("works", func() {
				brJob.RunOnVMAndSucceed(
					fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/backup --artifact-file %s --config %s",
						dbDumpPath, configPath))

				RunSQLCommand("UPDATE people SET NAME = 'New Person';", connection)

				brJob.RunOnVMAndSucceed(
					fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/restore --artifact-file %s --config %s",
						dbDumpPath, configPath))

				Expect(FetchSQLColumn("SELECT name FROM people;", connection)).To(ConsistOf("Old Person"))
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
							"port": %s,
							"database": "%s",
							"adapter": "mysql",
							"tls": {
								"cert": {
									"ca": "%s"
								}
							}
						}`,
							mysqlSslUsername,
							mysqlPassword,
							mysqlHostName,
							mysqlPort,
							databaseName,
							escapeNewLines(mysqlCaCert),
						)
					})

					It("works", func() {
						brJob.RunOnVMAndSucceed(
							fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/backup --artifact-file %s --config %s",
								dbDumpPath, configPath))

						RunSQLCommand("UPDATE people SET NAME = 'New Person';", connection)

						brJob.RunOnVMAndSucceed(
							fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/restore --artifact-file %s --config %s",
								dbDumpPath, configPath))

						Expect(FetchSQLColumn("SELECT name FROM people;", connection)).To(ConsistOf("Old Person"))
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
							"port": %s,
							"database": "%s",
							"adapter": "mysql",
							"tls": {
								"skip_host_verify": true,
								"cert": {
									"ca": "%s"
								}
							}
						}`,
							mysqlSslUsername,
							mysqlPassword,
							mysqlHostName,
							mysqlPort,
							databaseName,
							escapeNewLines(mysqlCaCert),
						)
					})

					It("works", func() {
						brJob.RunOnVMAndSucceed(
							fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/backup --artifact-file %s --config %s",
								dbDumpPath, configPath))

						RunSQLCommand("UPDATE people SET NAME = 'New Person';", connection)

						brJob.RunOnVMAndSucceed(
							fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/restore --artifact-file %s --config %s",
								dbDumpPath, configPath))

						Expect(FetchSQLColumn("SELECT name FROM people;", connection)).To(ConsistOf("Old Person"))
					})
				})
				Context("And the CA cert is malformed", func() {
					BeforeEach(func() {
						configJson = fmt.Sprintf(
							`{
									"username": "%s",
									"password": "%s",
									"host": "%s",
									"port": %s,
									"database": "%s",
									"adapter": "mysql",
									"tls": {
										"skip_host_verify": true,
										"cert": {
											"ca": "fooooooo"
										}
									}
								}`,
							mysqlSslUsername,
							mysqlPassword,
							mysqlHostName,
							mysqlPort,
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
						"port": %s,
						"database": "%s",
						"adapter": "mysql",
						"tls": {
							"cert": {
								"ca": "%s"
							}
						}
					}`,
						mysqlNonSslUsername,
						mysqlPassword,
						mysqlHostName,
						mysqlPort,
						databaseName,
						escapeNewLines(mysqlCaCert),
					)
				})

				It("works", func() {
					brJob.RunOnVMAndSucceed(
						fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/backup --artifact-file %s --config %s",
							dbDumpPath, configPath))

					RunSQLCommand("UPDATE people SET NAME = 'New Person';", connection)

					brJob.RunOnVMAndSucceed(
						fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/restore --artifact-file %s --config %s",
							dbDumpPath, configPath))

					Expect(FetchSQLColumn("SELECT name FROM people;", connection)).To(ConsistOf("Old Person"))
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
						"port": %s,
						"database": "%s",
						"adapter": "mysql",
						"tls": {
							"skip_host_verify": true,
							"cert": {
								"ca": "%s"
							}
						}
					}`,
						mysqlNonSslUsername,
						mysqlPassword,
						mysqlHostName,
						mysqlPort,
						databaseName,
						escapeNewLines(mysqlCaCert),
					)
				})

				It("works", func() {
					brJob.RunOnVMAndSucceed(
						fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/backup --artifact-file %s --config %s",
							dbDumpPath, configPath))

					RunSQLCommand("UPDATE people SET NAME = 'New Person';", connection)

					brJob.RunOnVMAndSucceed(
						fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/restore --artifact-file %s --config %s",
							dbDumpPath, configPath))

					Expect(FetchSQLColumn("SELECT name FROM people;", connection)).To(ConsistOf("Old Person"))
				})
			})
			Context("and the CA cert is malformed", func() {
				BeforeEach(func() {
					configJson = fmt.Sprintf(
						`{
								"username": "%s",
								"password": "%s",
								"host": "%s",
								"port": %s,
								"database": "%s",
								"adapter": "mysql",
								"tls": {
									"skip_host_verify": true,
									"cert": {
										"ca": "fooooooo"
									}
								}
							}`,
						mysqlNonSslUsername,
						mysqlPassword,
						mysqlHostName,
						mysqlPort,
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
