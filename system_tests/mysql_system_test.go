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
	"fmt"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("mysql", func() {
	var dbDumpPath string
	var configPath string
	var databaseName string
	var dbJob, brJob JobInstance

	BeforeEach(func() {
		configPath = "/tmp/config.json" + strconv.FormatInt(time.Now().Unix(), 10)
		dbDumpPath = "/tmp/sql_dump" + strconv.FormatInt(time.Now().Unix(), 10)
		databaseName = "db" + strconv.FormatInt(time.Now().Unix(), 10)
	})

	Context("when the mysql server version matches", func() {
		BeforeEach(func() {
			brJob = JobInstance{
				deployment:    "mysql-dev",
				instance:      "database-backup-restorer",
				instanceIndex: "0",
			}

			dbJob = JobInstance{
				deployment:    "mysql-dev",
				instance:      "mysql",
				instanceIndex: "0",
			}

			dbJob.runOnVMAndSucceed(fmt.Sprintf(`echo 'CREATE DATABASE %s;' | /var/vcap/packages/mariadb/bin/mysql -u root -h localhost --password='%s'`, databaseName, MustHaveEnv("MYSQL_PASSWORD")))
			dbJob.runMysqlSqlCommand("CREATE TABLE people (name varchar(255));", databaseName)
			dbJob.runMysqlSqlCommand("INSERT INTO people VALUES ('Derik');", databaseName)
			dbJob.runMysqlSqlCommand("CREATE TABLE places (name varchar(255));", databaseName)
			dbJob.runMysqlSqlCommand("INSERT INTO places VALUES ('London');", databaseName)

			configJson := fmt.Sprintf(
				`{"username":"root","password":"%s","host":"%s","port":3306,"database":"%s","adapter":"mysql"}`,
				MustHaveEnv("MYSQL_PASSWORD"),
				dbJob.getIPOfInstance(),
				databaseName,
			)
			brJob.RunOnInstance(fmt.Sprintf("echo '%s' > %s", configJson, configPath))
		})

		AfterEach(func() {
			dbJob.runOnVMAndSucceed(fmt.Sprintf(`echo 'DROP DATABASE %s;' | /var/vcap/packages/mariadb/bin/mysql -u root -h localhost --password='%s'`, databaseName, MustHaveEnv("MYSQL_PASSWORD")))
			brJob.RunOnInstance(fmt.Sprintf("rm -rf %s %s", configPath, dbDumpPath))
		})

		It("backs up and restores the database", func() {
			backupSession := brJob.RunOnInstance(fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/backup --artifact-file %s --config %s", dbDumpPath, configPath))
			Expect(backupSession).To(gexec.Exit(0))

			dbJob.runMysqlSqlCommand("UPDATE people SET NAME = 'Dave';", databaseName)
			dbJob.runMysqlSqlCommand("UPDATE places SET NAME = 'Rome';", databaseName)

			restoreSession := brJob.RunOnInstance(fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/restore --artifact-file %s --config %s", dbDumpPath, configPath))
			Expect(restoreSession).To(gexec.Exit(0))

			Expect(dbJob.runMysqlSqlCommand("SELECT name FROM people;", databaseName)).To(gbytes.Say("Derik"))
			Expect(dbJob.runMysqlSqlCommand("SELECT name FROM people;", databaseName)).NotTo(gbytes.Say("Dave"))
			Expect(dbJob.runMysqlSqlCommand("SELECT name FROM places;", databaseName)).To(gbytes.Say("London"))
			Expect(dbJob.runMysqlSqlCommand("SELECT name FROM places;", databaseName)).NotTo(gbytes.Say("Rome"))
		})

		Context("and 'tables' are specified in config", func() {
			BeforeEach(func() {
				configJson := fmt.Sprintf(
					`{"username":"root","password":"%s","host":"%s","port":3306,"database":"%s","adapter":"mysql","tables":["people"]}`,
					MustHaveEnv("MYSQL_PASSWORD"),
					dbJob.getIPOfInstance(),
					databaseName,
				)
				brJob.RunOnInstance(fmt.Sprintf("echo '%s' > %s", configJson, configPath))
			})

			It("backs up and restores only the specified tables", func() {
				backupSession := brJob.RunOnInstance(fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/backup --artifact-file %s --config %s", dbDumpPath, configPath))
				Expect(backupSession).To(gexec.Exit(0))

				dbJob.runMysqlSqlCommand("UPDATE people SET NAME = 'Dave';", databaseName)
				dbJob.runMysqlSqlCommand("UPDATE places SET NAME = 'Rome';", databaseName)

				testSession := brJob.RunOnInstance(fmt.Sprintf("cat %s", dbDumpPath))
				Expect(testSession).To(gexec.Exit(0))

				restoreSession := brJob.RunOnInstance(fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/restore --artifact-file %s --config %s", dbDumpPath, configPath))
				Expect(restoreSession).To(gexec.Exit(0))

				Expect(dbJob.runMysqlSqlCommand("SELECT name FROM people;", databaseName)).To(gbytes.Say("Derik"))
				Expect(dbJob.runMysqlSqlCommand("SELECT name FROM people;", databaseName)).NotTo(gbytes.Say("Dave"))
				Expect(dbJob.runMysqlSqlCommand("SELECT name FROM places;", databaseName)).To(gbytes.Say("Rome"))
				Expect(dbJob.runMysqlSqlCommand("SELECT name FROM places;", databaseName)).NotTo(gbytes.Say("London"))
			})
		})
	})

	Context("when the mysql server version doesn't match", func() {
		BeforeEach(func() {
			brJob = JobInstance{
				deployment:    "mysql-dev-old",
				instance:      "database-backup-restorer",
				instanceIndex: "0",
			}

			dbJob = JobInstance{
				deployment:    "mysql-dev-old",
				instance:      "mysql",
				instanceIndex: "0",
			}

			ip := dbJob.getIPOfInstance()
			configJson := fmt.Sprintf(
				`{"username":"root","password":"%s","host":"%s","port":3306,"database":"%s","adapter":"mysql"}`,
				MustHaveEnv("MYSQL_PASSWORD"),
				ip,
				databaseName,
			)

			brJob.RunOnInstance(fmt.Sprintf("echo '%s' > %s", configJson, configPath))

		})

		It("fails with a helpful message", func() {
			backupSession := brJob.RunOnInstance(fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/backup --artifact-file %s --config %s", dbDumpPath, configPath))
			Expect(backupSession).To(gexec.Exit(1))
			Expect(backupSession).To(gbytes.Say("Version mismatch"))
		})
	})
})
