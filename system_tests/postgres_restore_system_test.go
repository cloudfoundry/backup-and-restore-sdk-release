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

var _ = Describe("postgres-restore", func() {
	var databaseName string
	var dbDumpPath string
	var configPath string

	var dbJob JobInstance

	BeforeEach(func() {
		dbJob = JobInstance{
			deployment:    "postgres-dev",
			instance:      "postgres",
			instanceIndex: "0",
		}

		databaseName = "db" + strconv.FormatInt(time.Now().Unix(), 10)
		dbJob.runOnVMAndSucceed(fmt.Sprintf(`/var/vcap/packages/postgres-9.4/bin/createdb -U vcap "%s"`, databaseName))
		dbJob.runPostgresSqlCommand("CREATE TABLE people (name varchar);", databaseName)
		dbJob.runPostgresSqlCommand("INSERT INTO people VALUES ('Derik');", databaseName)

		configPath = "/tmp/config.json" + strconv.FormatInt(time.Now().Unix(), 10)
		dbDumpPath = "/tmp/sql_dump" + strconv.FormatInt(time.Now().Unix(), 10)

		configJson := fmt.Sprintf(
			`{"username":"vcap","password":"%s","host":"localhost","port":5432,"database":"%s","adapter":"postgres"}`,
			MustHaveEnv("POSTGRES_PASSWORD"),
			databaseName,
		)
		dbJob.runOnVMAndSucceed(fmt.Sprintf("echo '%s' > %s", configJson, configPath))
	})

	AfterEach(func() {
		dbJob.runOnVMAndSucceed(fmt.Sprintf(`/var/vcap/packages/postgres-9.4/bin/dropdb -U vcap "%s"`, databaseName))
		dbJob.runOnVMAndSucceed(fmt.Sprintf("rm -rf %s %s", configPath, dbDumpPath))
	})

	Context("database-backup-restorer is colocated with Postgres", func() {
		It("restores the Postgres database", func() {
			dbJob.runOnVMAndSucceed(fmt.Sprintf(`/var/vcap/jobs/database-backup-restorer/bin/backup --config %s --artifact-file %s`, configPath, dbDumpPath))

			dbJob.runPostgresSqlCommand("UPDATE people SET NAME = 'Dave';", databaseName)

			dbJob.runOnVMAndSucceed(fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/restore --config %s --artifact-file %s", configPath, dbDumpPath))

			Expect(dbJob.runPostgresSqlCommand("SELECT name FROM people;", databaseName)).To(gbytes.Say("Derik"))
			Expect(dbJob.runPostgresSqlCommand("SELECT name FROM people;", databaseName)).NotTo(gbytes.Say("Dave"))
		})
	})
})
