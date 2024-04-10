package postgresql_mutual_tls_test

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	"os"

	_ "github.com/lib/pq"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	. "database-backup-restore/system_tests/utils"
)

var _ = Describe("postgres with mutual tls", func() {
	var (
		dbDumpPath   string
		configPath   string
		databaseName string
		configJson   string

		pgConnection *PostgresConnection

		postgresClientCertPath string
		postgresClientKeyPath  string
	)

	BeforeEach(func() {
		disambiguationString := DisambiguationString()
		configPath = "/tmp/config" + disambiguationString
		dbDumpPath = "/tmp/artifact" + disambiguationString
		databaseName = "db" + disambiguationString

		postgresClientCertPath = writeToFile(postgresClientCert)
		postgresClientKeyPath = writeToFile(postgresClientKey)

		pgConnection = NewMutualTlsPostgresConnection(
			postgresHostName,
			postgresPort,
			postgresUsername,
			postgresPassword,
			postgresClientCertPath,
			postgresClientKeyPath,
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

		Expect(os.Remove(postgresClientCertPath)).To(Succeed())
		Expect(os.Remove(postgresClientKeyPath)).To(Succeed())

		brJob.RunOnVMAndSucceed(fmt.Sprintf("sudo rm -rf %s %s", configPath, dbDumpPath))
	})

	JustBeforeEach(func() {
		brJob.RunOnVMAndSucceed(fmt.Sprintf("echo '%s' > %s", configJson, configPath))
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

		It("does not work", func() {
			Expect(brJob.RunOnInstance("/var/vcap/jobs/database-backup-restorer/bin/backup",
				"--artifact-file", dbDumpPath, "--config", configPath)).To(gexec.Exit(1))
		})
	})

	Context("when TLS info is provided in the config", func() {
		Context("and host verification is not skipped", func() {
			BeforeEach(func() {
				maybeSkipTLSVerifyIdentityTests()
			})

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
						postgresUsername,
						postgresPassword,
						postgresHostName,
						postgresPort,
						databaseName,
						EscapeNewLines(postgresCaCert),
						EscapeNewLines(postgresClientCert),
						EscapeNewLines(postgresClientKey),
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
						postgresUsername,
						postgresPassword,
						postgresHostName,
						postgresPort,
						databaseName,
						EscapeNewLines(postgresCaCert),
						EscapeNewLines(postgresClientCert),
						EscapeNewLines(postgresClientKey),
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
						postgresUsername,
						postgresPassword,
						postgresHostName,
						postgresPort,
						databaseName,
						EscapeNewLines(postgresCaCert),
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
						postgresUsername,
						postgresPassword,
						postgresHostName,
						postgresPort,
						databaseName,
						EscapeNewLines(postgresCaCert),
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

func writeToFile(contents string) string {
	file, err := os.CreateTemp("", "")
	Expect(err).NotTo(HaveOccurred())
	path := file.Name()
	Expect(os.WriteFile(path, []byte(contents), 0777)).To(Succeed())
	return path
}
