package postgresql_mutual_tls_test

import (
	. "database-backup-restore/system_tests/utils"

	"fmt"
	"os"

	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"io/ioutil"

	_ "github.com/lib/pq"
)

var _ = Describe("postgres with mutual tls", func() {
	var dbDumpPath string
	var configPath string
	var databaseName string
	var configJson string

	var pgConnection *PostgresConnection

	var postgresHostName string
	var postgresUsername string
	var postgresPassword string
	var postgresPort int
	var postgresCaCert string
	var postgresClientCert string
	var postgresClientKey string
	var postgresClientCertPath string
	var postgresClientKeyPath string

	var brJob JobInstance

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
		postgresClientCert = os.Getenv("POSTGRES_CLIENT_CERT")
		postgresClientKey = os.Getenv("POSTGRES_CLIENT_KEY")
	})

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

		os.Remove(postgresClientCertPath)
		os.Remove(postgresClientKeyPath)

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
	file, err := ioutil.TempFile("", "")
	Expect(err).NotTo(HaveOccurred())
	path := file.Name()
	ioutil.WriteFile(path, []byte(contents), 0777)
	return path
}
