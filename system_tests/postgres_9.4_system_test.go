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

package system_tests

import (
	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"fmt"
	"time"
)

var _ = Describe("postgres-9.4", func() {
	var databaseName string
	var dbDumpPath string
	var configPath string
	postgresPackage := "postgres-9.4"

	var brJob, dbJob JobInstance

	BeforeEach(func() {
		brJob = JobInstance{
			deployment:    "postgres-9.4-dev",
			instance:      "database-backup-restorer",
			instanceIndex: "0",
		}

		dbJob = JobInstance{
			deployment:    "postgres-9.4-dev",
			instance:      "postgres",
			instanceIndex: "0",
		}

		databaseName = "db" + strconv.FormatInt(time.Now().Unix(), 10)
		dbJob.runPostgresSqlCommand("ALTER ROLE test_user SUPERUSER;", "postgres", postgresPackage)
		dbJob.runOnVMAndSucceed(
			fmt.Sprintf(`/var/vcap/packages/%s/bin/createdb -U test_user "%s"`,
				postgresPackage, databaseName))
		dbJob.runPostgresSqlCommand("CREATE TABLE people (name varchar);", databaseName, postgresPackage)
		dbJob.runPostgresSqlCommand("INSERT INTO people VALUES ('Derik');", databaseName, postgresPackage)
		dbJob.runPostgresSqlCommand("CREATE TABLE places (name varchar);", databaseName, postgresPackage)
		dbJob.runPostgresSqlCommand("INSERT INTO places VALUES ('London');", databaseName, postgresPackage)

		configPath = "/tmp/config.json" + strconv.FormatInt(time.Now().Unix(), 10)
		dbDumpPath = "/tmp/sql_dump" + strconv.FormatInt(time.Now().Unix(), 10)

		ip := dbJob.getIPOfInstance()
		configJson := fmt.Sprintf(
			`{"username":"test_user","password":"%s","host":"%s","port":5432,
				"database":"%s","adapter":"postgres"}`,
			MustHaveEnv("POSTGRES_PASSWORD"),
			ip,
			databaseName,
		)

		brJob.runOnVMAndSucceed(fmt.Sprintf("echo '%s' > %s", configJson, configPath))
	})

	AfterEach(func() {
		dbJob.runOnVMAndSucceed(fmt.Sprintf(
			`/var/vcap/packages/%s/bin/dropdb -U test_user "%s"`, postgresPackage, databaseName))
		brJob.runOnVMAndSucceed(fmt.Sprintf("rm -rf %s %s", configPath, dbDumpPath))
	})

	Context("database-backup-restorer is on its own instance", func() {
		It("restores the Postgres database", func() {
			brJob.runOnVMAndSucceed(
				fmt.Sprintf(`/var/vcap/jobs/database-backup-restorer/bin/backup --config %s --artifact-file %s`,
					configPath, dbDumpPath))
			brJob.runOnVMAndSucceed(fmt.Sprintf("ls -l %s", dbDumpPath))

			dbJob.runPostgresSqlCommand("UPDATE people SET NAME = 'Dave';", databaseName, postgresPackage)
			dbJob.runPostgresSqlCommand("UPDATE places SET NAME = 'Rome';", databaseName, postgresPackage)

			brJob.runOnVMAndSucceed(
				fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/restore --config %s --artifact-file %s",
					configPath, dbDumpPath))

			Expect(dbJob.runPostgresSqlCommand("SELECT name FROM people;", databaseName, postgresPackage)).To(gbytes.Say("Derik"))
			Expect(dbJob.runPostgresSqlCommand("SELECT name FROM people;", databaseName, postgresPackage)).NotTo(gbytes.Say("Dave"))
			Expect(dbJob.runPostgresSqlCommand("SELECT name FROM places;", databaseName, postgresPackage)).To(gbytes.Say("London"))
			Expect(dbJob.runPostgresSqlCommand("SELECT name FROM places;", databaseName, postgresPackage)).NotTo(gbytes.Say("Rome"))
		})
	})
})
