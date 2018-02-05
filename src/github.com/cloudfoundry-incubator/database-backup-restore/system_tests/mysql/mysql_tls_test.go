package mysql

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	_ "github.com/go-sql-driver/mysql"

	"strings"

	. "github.com/cloudfoundry-incubator/database-backup-restore/system_tests/utils"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("mysql with tls", func() {
	var dbDumpPath string
	var configPath string
	var databaseName string
	var sslUser string
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

		sslUser = "ssl_user_" + DisambiguationStringOfLength(6)
		RunSQLCommand(fmt.Sprintf(
			"CREATE USER '%s' IDENTIFIED BY '%s';",
			sslUser, mysqlPassword), connection)
	})

	AfterEach(func() {
		RunSQLCommand(fmt.Sprintf("DROP USER '%s';", sslUser), connection)
	})

	JustBeforeEach(func() {
		brJob.RunOnVMAndSucceed(fmt.Sprintf("echo '%s' > %s", configJson, configPath))
	})

	Context("when the db user requires TLS", func() {
		BeforeEach(func() {
			RunSQLCommand(fmt.Sprintf("GRANT ALL PRIVILEGES ON %s.* TO %s REQUIRE SSL;", databaseName, sslUser), connection)
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
					sslUser,
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
							sslUser,
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
							sslUser,
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
							sslUser,
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

	Context("the user does not require TLS", func() {
		Context("correct CA cert is provided", func() {
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
					mysqlUsername,
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
								"cert": {
									"ca": "fooooooo"
								}
							}
						}`,
					mysqlUsername,
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

func escapeNewLines(txt string) string {
	return strings.Replace(txt, "\n", "\\n", -1)
}
