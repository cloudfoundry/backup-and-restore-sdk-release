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

package postgresql

import (
	"fmt"
	"os"
	"strconv"

	_ "github.com/lib/pq"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	. "database-backup-restore/system_tests/utils"
)

var (
	pgConnection *PostgresConnection

	postgresHostName string
	postgresUsername string
	postgresPassword string
	postgresPort     int

	brJob JobInstance
)

var _ = Describe("postgres", func() {
	var (
		databaseName string
		dbDumpPath   string
		configPath   string
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
			false,
		)

		pgConnection.OpenSuccessfully("postgres")
		pgConnection.RunSQLCommand("CREATE DATABASE " + databaseName)
		pgConnection.SwitchToDb(databaseName)
		pgConnection.RunSQLCommand("CREATE TABLE people (name varchar(255) unique);")
		pgConnection.RunSQLCommand("INSERT INTO people VALUES ('Old Person');")
		pgConnection.RunSQLCommand("CREATE TABLE places (name varchar);")
		pgConnection.RunSQLCommand("INSERT INTO places VALUES ('Old Place');")
	})

	AfterEach(func() {
		pgConnection.SwitchToDb("postgres")
		pgConnection.RunSQLCommand("DROP DATABASE " + databaseName)
		pgConnection.Close()

		brJob.RunOnVMAndSucceed(fmt.Sprintf("sudo rm -rf %s %s", configPath, dbDumpPath))
	})

	Context("database backup is successful", func() {
		BeforeEach(func() {
			configJson := fmt.Sprintf(
				`{
					"username": "%s",
					"password": "%s",
					"host": "%s",
					"port": 5432,
					"database": "%s",
					"adapter": "postgres"
				}`,
				postgresUsername,
				postgresPassword,
				postgresHostName,
				databaseName,
			)
			brJob.RunOnVMAndSucceed(fmt.Sprintf("echo '%s' > %s", configJson, configPath))
			brJob.RunOnVMAndSucceed(
				fmt.Sprintf(`/var/vcap/jobs/database-backup-restorer/bin/backup --config %s --artifact-file %s`,
					configPath, dbDumpPath))
			brJob.RunOnVMAndSucceed(fmt.Sprintf("ls -l %s", dbDumpPath))
		})

		Context("when tables exist", func() {
			BeforeEach(func() {
				pgConnection.RunSQLCommand("UPDATE people SET NAME = 'New Person';")
				pgConnection.RunSQLCommand("UPDATE places SET NAME = 'New Place';")

			})

			It("restores the Postgres database", func() {
				brJob.RunOnVMAndSucceed(
					fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/restore --config %s --artifact-file %s",
						configPath, dbDumpPath))

				Expect(pgConnection.FetchSQLColumn("SELECT name FROM people;")).
					To(ConsistOf("Old Person"))
				Expect(pgConnection.FetchSQLColumn("SELECT name FROM people;")).
					NotTo(ConsistOf("New Person"))
				Expect(pgConnection.FetchSQLColumn("SELECT name FROM places;")).
					To(ConsistOf("Old Place"))
				Expect(pgConnection.FetchSQLColumn("SELECT name FROM places;")).
					NotTo(ConsistOf("New Place"))
			})

			Context("when the restore fails due to an SQL error", func() {
				BeforeEach(func() {
					pgConnection.RunSQLCommand("CREATE TABLE children (name varchar(255) REFERENCES people(name));")
				})

				It("the restore is atomic and does not contain any new data", func() {
					session := brJob.RunOnInstance(
						fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/restore --config %s --artifact-file %s",
							configPath, dbDumpPath))

					Expect(pgConnection.FetchSQLColumn("SELECT name FROM people;")).
						To(ConsistOf("New Person"))
					Expect(pgConnection.FetchSQLColumn("SELECT name FROM people;")).
						NotTo(ConsistOf("Old Person"))
					Expect(pgConnection.FetchSQLColumn("SELECT name FROM places;")).
						To(ConsistOf("New Place"))
					Expect(pgConnection.FetchSQLColumn("SELECT name FROM places;")).
						NotTo(ConsistOf("Old Place"))
					Expect(session.ExitCode()).To(Equal(1))
				})
			})
		})

		Context("when tables do not exist", func() {
			It("restores the tables successfully", func() {
				pgConnection.RunSQLCommand("DROP TABLE people;")

				brJob.RunOnVMAndSucceed(
					fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/restore --config %s --artifact-file %s",
						configPath, dbDumpPath))

				Expect(pgConnection.FetchSQLColumn("SELECT name FROM people;")).
					To(ConsistOf("Old Person"))
			})
		})
	})

	Context("and 'tables' are specified in config", func() {
		BeforeEach(func() {
			configJson := fmt.Sprintf(
				`{
					"username": "%s",
					"password": "%s",
					"host": "%s",
					"port": 5432,
					"database": "%s",
					"adapter": "postgres",
					"tables": ["people"]
				}`,
				postgresUsername,
				postgresPassword,
				postgresHostName,
				databaseName,
			)
			brJob.RunOnVMAndSucceed(fmt.Sprintf("echo '%s' > %s", configJson, configPath))
		})

		It("backs up and restores only the specified tables", func() {
			brJob.RunOnVMAndSucceed(fmt.Sprintf(
				"/var/vcap/jobs/database-backup-restorer/bin/backup --artifact-file %s --config %s",
				dbDumpPath,
				configPath))

			pgConnection.RunSQLCommand("UPDATE people SET NAME = 'New Person';")
			pgConnection.RunSQLCommand("UPDATE places SET NAME = 'New Place';")

			brJob.RunOnVMAndSucceed(fmt.Sprintf("cat %s", dbDumpPath))

			brJob.RunOnVMAndSucceed(fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/restore --artifact-file %s --config %s", dbDumpPath, configPath))

			Expect(pgConnection.FetchSQLColumn("SELECT name FROM people;")).
				To(ConsistOf("Old Person"))
			Expect(pgConnection.FetchSQLColumn("SELECT name FROM people;")).
				NotTo(ConsistOf("New Person"))
			Expect(pgConnection.FetchSQLColumn("SELECT name FROM places;")).
				To(ConsistOf("New Place"))
			Expect(pgConnection.FetchSQLColumn("SELECT name FROM places;")).
				NotTo(ConsistOf("Old Place"))
		})
	})

	Context("and 'tables' are specified in config, with a non-existent table", func() {
		BeforeEach(func() {
			configJson := fmt.Sprintf(
				`{
					"username": "%s",
					"password": "%s",
					"host": "%s",
					"port": 5432,
					"database": "%s",
					"adapter": "postgres",
					"tables": ["people", "lizards"]
				}`,
				postgresUsername,
				postgresPassword,
				postgresHostName,
				databaseName,
			)
			brJob.RunOnVMAndSucceed(fmt.Sprintf("echo '%s' > %s", configJson, configPath))
		})

		It("raises an error about the non-existent tables", func() {
			session := brJob.RunOnInstance(fmt.Sprintf(
				"/var/vcap/jobs/database-backup-restorer/bin/backup --artifact-file %s --config %s",
				dbDumpPath,
				configPath))
			Expect(session.ExitCode()).NotTo(BeZero())
			Expect(session).To(gbytes.Say(`can't find specified table\(s\): lizards`))
		})
	})
})
