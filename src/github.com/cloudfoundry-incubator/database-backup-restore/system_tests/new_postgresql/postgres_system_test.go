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

package new_postgresql

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"fmt"

	. "github.com/cloudfoundry-incubator/database-backup-restore/system_tests/utils"
)

var _ = Describe("postgres", func() {
	postgresPackage, postgresDeployment := "postgres-9.4", "postgres-9.4-dev"
	var databaseName string
	var dbDumpPath string
	var configPath string

	var dbHostName string

	var brJob, dbJob JobInstance

	const testUser = "test_user"

	BeforeEach(func() {

		dbHostName = MustHaveEnv("POSTGRES_HOSTNAME")

		brJob = JobInstance{
			Deployment:    postgresDeployment,
			Instance:      "database-backup-restorer",
			InstanceIndex: "0",
		}

		dbJob = JobInstance{
			Deployment:    postgresDeployment,
			Instance:      "postgres",
			InstanceIndex: "0",
		}

		disambiguationString := DisambiguationString()
		databaseName = "db" + disambiguationString
		configPath = "/tmp/config" + disambiguationString
		dbDumpPath = "/tmp/artifact" + disambiguationString

		dbJob.RunOnVMAndSucceed(
			fmt.Sprintf(`/var/vcap/packages/%s/bin/createdb -U vcap "%s"`,
				postgresPackage, databaseName))
		dbJob.RunPostgresSqlCommand("CREATE TABLE people (name varchar);", databaseName, testUser, postgresPackage)
		dbJob.RunPostgresSqlCommand("INSERT INTO people VALUES ('Old Person');", databaseName, testUser, postgresPackage)
		dbJob.RunPostgresSqlCommand("CREATE TABLE places (name varchar);", databaseName, testUser, postgresPackage)
		dbJob.RunPostgresSqlCommand("INSERT INTO places VALUES ('Old Place');", databaseName, testUser, postgresPackage)

		configJson := fmt.Sprintf(
			`{"username":"test_user","password":"%s","host":"%s","port":5432,
						"database":"%s","adapter":"postgres"}`,
			MustHaveEnv("POSTGRES_PASSWORD"),
			dbHostName,
			databaseName,
		)
		brJob.RunOnVMAndSucceed(fmt.Sprintf("echo '%s' > %s", configJson, configPath))
	})

	AfterEach(func() {
		dbJob.RunOnVMAndSucceed(fmt.Sprintf(
			`/var/vcap/packages/%s/bin/dropdb -U vcap "%s"`, postgresPackage, databaseName))
		brJob.RunOnVMAndSucceed(fmt.Sprintf("rm -rf %s %s", configPath, dbDumpPath))
	})

	It("backs up the Postgres database", func() {
		brJob.RunOnVMAndSucceed(
			fmt.Sprintf(`/var/vcap/jobs/database-backup-restorer/bin/backup --config %s --artifact-file %s`,
				configPath, dbDumpPath))
		brJob.RunOnVMAndSucceed(fmt.Sprintf("ls -l %s", dbDumpPath))

		dbJob.RunPostgresSqlCommand("UPDATE people SET NAME = 'New Person';", databaseName, testUser, postgresPackage)
		dbJob.RunPostgresSqlCommand("UPDATE places SET NAME = 'New Place';", databaseName, testUser, postgresPackage)

		brJob.RunOnVMAndSucceed(
			fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/restore --config %s --artifact-file %s",
				configPath, dbDumpPath))

		Expect(dbJob.RunPostgresSqlCommand("SELECT name FROM people;", databaseName, testUser, postgresPackage)).
			To(gbytes.Say("Old Person"))
		Expect(dbJob.RunPostgresSqlCommand("SELECT name FROM people;", databaseName, testUser, postgresPackage)).
			NotTo(gbytes.Say("New Person"))
		Expect(dbJob.RunPostgresSqlCommand("SELECT name FROM places;", databaseName, testUser, postgresPackage)).
			To(gbytes.Say("Old Place"))
		Expect(dbJob.RunPostgresSqlCommand("SELECT name FROM places;", databaseName, testUser, postgresPackage)).
			NotTo(gbytes.Say("New Place"))
	})

	Context("and 'tables' are specified in config", func() {
		BeforeEach(func() {
			configJson := fmt.Sprintf(
				`{"username":"test_user","password":"%s","host":"%s","port":5432,
							"database":"%s","adapter":"postgres", "tables":["people"]}`,
				MustHaveEnv("POSTGRES_PASSWORD"),
				dbHostName,
				databaseName,
			)
			brJob.RunOnVMAndSucceed(fmt.Sprintf("echo '%s' > %s", configJson, configPath))
		})

		It("backs up and restores only the specified tables", func() {
			brJob.RunOnVMAndSucceed(fmt.Sprintf(
				"/var/vcap/jobs/database-backup-restorer/bin/backup --artifact-file %s --config %s",
				dbDumpPath,
				configPath))

			dbJob.RunPostgresSqlCommand("UPDATE people SET NAME = 'New Person';", databaseName, testUser, postgresPackage)
			dbJob.RunPostgresSqlCommand("UPDATE places SET NAME = 'New Place';", databaseName, testUser, postgresPackage)

			brJob.RunOnVMAndSucceed(fmt.Sprintf("cat %s", dbDumpPath))

			brJob.RunOnVMAndSucceed(fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/restore --artifact-file %s --config %s", dbDumpPath, configPath))

			Expect(dbJob.RunPostgresSqlCommand("SELECT name FROM people;", databaseName, testUser, postgresPackage)).
				To(gbytes.Say("Old Person"))
			Expect(dbJob.RunPostgresSqlCommand("SELECT name FROM people;", databaseName, testUser, postgresPackage)).
				NotTo(gbytes.Say("New Person"))
			Expect(dbJob.RunPostgresSqlCommand("SELECT name FROM places;", databaseName, testUser, postgresPackage)).
				To(gbytes.Say("New Place"))
			Expect(dbJob.RunPostgresSqlCommand("SELECT name FROM places;", databaseName, testUser, postgresPackage)).
				NotTo(gbytes.Say("Old Place"))
		})
	})

	Context("and 'tables' are specified in config, with a non-existent table", func() {
		BeforeEach(func() {
			configJson := fmt.Sprintf(
				`{"username":"test_user","password":"%s","host":"%s","port":5432,
							"database":"%s","adapter":"postgres", "tables":["people", "lizards"]}`,
				MustHaveEnv("POSTGRES_PASSWORD"),
				dbHostName,
				databaseName,
			)
			brJob.RunOnVMAndSucceed(fmt.Sprintf("echo '%s' > %s", configJson, configPath))
		})

		It("backs up and restores only the specified tables", func() {
			session := brJob.RunOnInstance(fmt.Sprintf(
				"/var/vcap/jobs/database-backup-restorer/bin/backup --artifact-file %s --config %s",
				dbDumpPath,
				configPath))
			Expect(session.ExitCode()).NotTo(BeZero())
			Expect(session).To(gbytes.Say(`can't find specified table\(s\): lizards`))
		})
	})
})
