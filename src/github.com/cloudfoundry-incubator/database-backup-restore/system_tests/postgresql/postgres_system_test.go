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
	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"fmt"
	"time"
)

func PostgresTests(postgresPackage, postgresDeployment string) func() {
	return func() {
		var databaseName string
		var dbDumpPath string
		var configPath string

		var brJob, dbJob JobInstance

		const testUser = "test_user"

		BeforeEach(func() {
			brJob = JobInstance{
				deployment:    postgresDeployment,
				instance:      "database-backup-restorer",
				instanceIndex: "0",
			}

			dbJob = JobInstance{
				deployment:    postgresDeployment,
				instance:      "postgres",
				instanceIndex: "0",
			}

			databaseName = "db" + strconv.FormatInt(time.Now().Unix(), 10)
			dbJob.runOnVMAndSucceed(
				fmt.Sprintf(`/var/vcap/packages/%s/bin/createdb -U vcap "%s"`,
					postgresPackage, databaseName))
			dbJob.runPostgresSqlCommand("CREATE TABLE people (name varchar);", databaseName, testUser, postgresPackage)
			dbJob.runPostgresSqlCommand("INSERT INTO people VALUES ('Old Person');", databaseName, testUser, postgresPackage)
			dbJob.runPostgresSqlCommand("CREATE TABLE places (name varchar);", databaseName, testUser, postgresPackage)
			dbJob.runPostgresSqlCommand("INSERT INTO places VALUES ('Old Place');", databaseName, testUser, postgresPackage)

			configPath = "/tmp/config.json" + strconv.FormatInt(time.Now().Unix(), 10)
			dbDumpPath = "/tmp/sql_dump" + strconv.FormatInt(time.Now().Unix(), 10)

			configJson := fmt.Sprintf(
				`{"username":"test_user","password":"%s","host":"%s","port":5432,
					"database":"%s","adapter":"postgres"}`,
				MustHaveEnv("POSTGRES_PASSWORD"),
				dbJob.getIPOfInstance(),
				databaseName,
			)
			brJob.runOnVMAndSucceed(fmt.Sprintf("echo '%s' > %s", configJson, configPath))
		})

		AfterEach(func() {
			dbJob.runOnVMAndSucceed(fmt.Sprintf(
				`/var/vcap/packages/%s/bin/dropdb -U vcap "%s"`, postgresPackage, databaseName))
			brJob.runOnVMAndSucceed(fmt.Sprintf("rm -rf %s %s", configPath, dbDumpPath))
		})

		It("backs up the Postgres database", func() {
			brJob.runOnVMAndSucceed(
				fmt.Sprintf(`/var/vcap/jobs/database-backup-restorer/bin/backup --config %s --artifact-file %s`,
					configPath, dbDumpPath))
			brJob.runOnVMAndSucceed(fmt.Sprintf("ls -l %s", dbDumpPath))

			dbJob.runPostgresSqlCommand("UPDATE people SET NAME = 'New Person';", databaseName, testUser, postgresPackage)
			dbJob.runPostgresSqlCommand("UPDATE places SET NAME = 'New Place';", databaseName, testUser, postgresPackage)

			brJob.runOnVMAndSucceed(
				fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/restore --config %s --artifact-file %s",
					configPath, dbDumpPath))

			Expect(dbJob.runPostgresSqlCommand("SELECT name FROM people;", databaseName, testUser, postgresPackage)).
				To(gbytes.Say("Old Person"))
			Expect(dbJob.runPostgresSqlCommand("SELECT name FROM people;", databaseName, testUser, postgresPackage)).
				NotTo(gbytes.Say("New Person"))
			Expect(dbJob.runPostgresSqlCommand("SELECT name FROM places;", databaseName, testUser, postgresPackage)).
				To(gbytes.Say("Old Place"))
			Expect(dbJob.runPostgresSqlCommand("SELECT name FROM places;", databaseName, testUser, postgresPackage)).
				NotTo(gbytes.Say("New Place"))
		})

		Context("and 'tables' are specified in config", func() {
			BeforeEach(func() {
				configJson := fmt.Sprintf(
					`{"username":"test_user","password":"%s","host":"%s","port":5432,
						"database":"%s","adapter":"postgres", "tables":["people"]}`,
					MustHaveEnv("POSTGRES_PASSWORD"),
					dbJob.getIPOfInstance(),
					databaseName,
				)
				brJob.runOnVMAndSucceed(fmt.Sprintf("echo '%s' > %s", configJson, configPath))
			})

			It("backs up and restores only the specified tables", func() {
				brJob.runOnVMAndSucceed(fmt.Sprintf(
					"/var/vcap/jobs/database-backup-restorer/bin/backup --artifact-file %s --config %s",
					dbDumpPath,
					configPath))

				dbJob.runPostgresSqlCommand("UPDATE people SET NAME = 'New Person';", databaseName, testUser, postgresPackage)
				dbJob.runPostgresSqlCommand("UPDATE places SET NAME = 'New Place';", databaseName, testUser, postgresPackage)

				brJob.runOnVMAndSucceed(fmt.Sprintf("cat %s", dbDumpPath))

				brJob.runOnVMAndSucceed(fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/restore --artifact-file %s --config %s", dbDumpPath, configPath))

				Expect(dbJob.runPostgresSqlCommand("SELECT name FROM people;", databaseName, testUser, postgresPackage)).
					To(gbytes.Say("Old Person"))
				Expect(dbJob.runPostgresSqlCommand("SELECT name FROM people;", databaseName, testUser, postgresPackage)).
					NotTo(gbytes.Say("New Person"))
				Expect(dbJob.runPostgresSqlCommand("SELECT name FROM places;", databaseName, testUser, postgresPackage)).
					To(gbytes.Say("New Place"))
				Expect(dbJob.runPostgresSqlCommand("SELECT name FROM places;", databaseName, testUser, postgresPackage)).
					NotTo(gbytes.Say("Old Place"))
			})
		})

		Context("and 'tables' are specified in config, with a non-existent table", func() {
			BeforeEach(func() {
				configJson := fmt.Sprintf(
					`{"username":"test_user","password":"%s","host":"%s","port":5432,
						"database":"%s","adapter":"postgres", "tables":["people", "lizards"]}`,
					MustHaveEnv("POSTGRES_PASSWORD"),
					dbJob.getIPOfInstance(),
					databaseName,
				)
				brJob.runOnVMAndSucceed(fmt.Sprintf("echo '%s' > %s", configJson, configPath))
			})

			It("backs up and restores only the specified tables", func() {
				session := brJob.runOnInstance(fmt.Sprintf(
					"/var/vcap/jobs/database-backup-restorer/bin/backup --artifact-file %s --config %s",
					dbDumpPath,
					configPath))
				Expect(session.ExitCode()).NotTo(BeZero())
				Expect(session).To(gbytes.Say(`can't find specified table\(s\): lizards`))
			})
		})
	}
}

var _ = Describe("postgres", func() {
	Context("9.4", PostgresTests("postgres-9.4", "postgres-9.4-dev"))
	Context("9.6", PostgresTests("postgres-9.6.3", "postgres-9.6-dev"))
	Context("tls", PostgresTests("postgres-9.6.3", "postgres-with-tls-dev"))
})
