package postgresql_tls_test

import (
	"fmt"
	"os"
	"strconv"

	_ "github.com/lib/pq"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	. "github.com/cloudfoundry/backup-and-restore-sdk-release/src/database-backup-restore/system_tests/utils"
)

var _ = Describe("postgres with tls", func() {
	var (
		dbDumpPath   string
		configPath   string
		databaseName string
		configJson   string

		pgConnection *PostgresConnection

		postgresHostName string
		postgresUsername string
		postgresPassword string
		postgresPort     int
		postgresCaCert   string

		brJob JobInstance
	)

	BeforeSuite(func() {
		brJob = JobInstance{
			Deployment:    MustHaveEnv("SDK_DEPLOYMENT"),
			Instance:      MustHaveEnv("SDK_INSTANCE_GROUP"),
			InstanceIndex: "0",
		}

		postgresHostName = MustHaveEnv("POSTGRES_HOSTNAME")
		postgresPort, _ = strconv.Atoi(MustHaveEnv("POSTGRES_PORT"))
		postgresPassword = MustHaveEnv("POSTGRES_PASSWORD")
		postgresUsername = MustHaveEnv("POSTGRES_USERNAME")
		postgresCaCert = os.Getenv("POSTGRES_CA_CERT")
	})

	BeforeEach(func() {
		disambiguationString := DisambiguationString()
		configPath = "/tmp/config" + disambiguationString
		dbDumpPath = "/tmp/artifact" + disambiguationString
		databaseName = "db" + disambiguationString

		pgConnection = NewPostgresConnection(
			postgresHostName,
			postgresPort,
			postgresUsername,
			postgresPassword,
			os.Getenv("SSH_PROXY_HOST"),
			os.Getenv("SSH_PROXY_USER"),
			os.Getenv("SSH_PROXY_KEY_FILE"),
			true,
		)

		pgConnection.OpenSuccessfully("postgres")
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
		if os.Getenv("TEST_SSL_USER_REQUIRES_SSL") == "false" {
			return
		}
		It("can't connect with ssl disabled", func() {
			err := NewPostgresConnection(
				postgresHostName,
				postgresPort,
				postgresUsername,
				postgresPassword,
				os.Getenv("SSH_PROXY_HOST"),
				os.Getenv("SSH_PROXY_USER"),
				os.Getenv("SSH_PROXY_KEY_FILE"),
				false,
			).Open(databaseName)

			Expect(err).To(MatchError(MatchRegexp("no pg_hba.conf entry for host \".*\", user \"ssl_user\", database \".*\", SSL off")))
		})
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
				postgresUsername,
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
			BeforeEach(func() {
				maybeSkipTLSVerifyIdentityTests()
			})

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
						postgresUsername,
						postgresPassword,
						postgresHostName,
						postgresPort,
						databaseName,
						EscapeNewLines(postgresCaCert),
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
						postgresUsername,
						postgresPassword,
						postgresHostName,
						postgresPort,
						databaseName,
						EscapeNewLines(postgresCaCert),
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
						postgresUsername,
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
