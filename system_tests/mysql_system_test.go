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
			dbJob.runMysqlSqlCommand("INSERT INTO people VALUES ('Old Person');", databaseName)
			dbJob.runMysqlSqlCommand("CREATE TABLE places (name varchar(255));", databaseName)
			dbJob.runMysqlSqlCommand("INSERT INTO places VALUES ('Old Place');", databaseName)

			configJson := fmt.Sprintf(
				`{"username":"root","password":"%s","host":"%s","port":3306,"database":"%s","adapter":"mysql"}`,
				MustHaveEnv("MYSQL_PASSWORD"),
				dbJob.getIPOfInstance(),
				databaseName,
			)
			brJob.runOnVMAndSucceed(fmt.Sprintf("echo '%s' > %s", configJson, configPath))
		})

		AfterEach(func() {
			dbJob.runOnVMAndSucceed(fmt.Sprintf(`echo 'DROP DATABASE %s;' | /var/vcap/packages/mariadb/bin/mysql -u root -h localhost --password='%s'`, databaseName, MustHaveEnv("MYSQL_PASSWORD")))
			brJob.runOnVMAndSucceed(fmt.Sprintf("rm -rf %s %s", configPath, dbDumpPath))
		})

		It("backs up and restores the database", func() {
			brJob.runOnVMAndSucceed(fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/backup --artifact-file %s --config %s", dbDumpPath, configPath))

			dbJob.runMysqlSqlCommand("UPDATE people SET NAME = 'New Person';", databaseName)
			dbJob.runMysqlSqlCommand("UPDATE places SET NAME = 'New Place';", databaseName)

			brJob.runOnVMAndSucceed(fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/restore --artifact-file %s --config %s", dbDumpPath, configPath))

			Expect(dbJob.runMysqlSqlCommand("SELECT name FROM people;", databaseName)).To(gbytes.Say("Old Person"))
			Expect(dbJob.runMysqlSqlCommand("SELECT name FROM people;", databaseName)).NotTo(gbytes.Say("New Person"))
			Expect(dbJob.runMysqlSqlCommand("SELECT name FROM places;", databaseName)).To(gbytes.Say("Old Place"))
			Expect(dbJob.runMysqlSqlCommand("SELECT name FROM places;", databaseName)).NotTo(gbytes.Say("New Place"))
		})

		Context("and some existing 'tables' are specified in config", func() {
			BeforeEach(func() {
				configJson := fmt.Sprintf(
					`{"username":"root","password":"%s","host":"%s","port":3306,"database":"%s","adapter":"mysql","tables":["people"]}`,
					MustHaveEnv("MYSQL_PASSWORD"),
					dbJob.getIPOfInstance(),
					databaseName,
				)
				brJob.runOnVMAndSucceed(fmt.Sprintf("echo '%s' > %s", configJson, configPath))
			})

			It("backs up and restores only the specified tables", func() {
				brJob.runOnVMAndSucceed(fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/backup --artifact-file %s --config %s", dbDumpPath, configPath))

				dbJob.runMysqlSqlCommand("UPDATE people SET NAME = 'New Person';", databaseName)
				dbJob.runMysqlSqlCommand("UPDATE places SET NAME = 'New Place';", databaseName)

				brJob.runOnVMAndSucceed(fmt.Sprintf("cat %s", dbDumpPath))

				restoreSession := brJob.runOnVMAndSucceed(fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/restore --artifact-file %s --config %s", dbDumpPath, configPath))

				Expect(restoreSession).To(gbytes.Say("CREATE TABLE `people`"))

				Expect(dbJob.runMysqlSqlCommand("SELECT name FROM people;", databaseName)).To(gbytes.Say("Old Person"))
				Expect(dbJob.runMysqlSqlCommand("SELECT name FROM people;", databaseName)).NotTo(gbytes.Say("New Person"))
				Expect(dbJob.runMysqlSqlCommand("SELECT name FROM places;", databaseName)).To(gbytes.Say("New Place"))
				Expect(dbJob.runMysqlSqlCommand("SELECT name FROM places;", databaseName)).NotTo(gbytes.Say("Old Place"))
			})
		})

		Context("and 'tables' are specified in config only some of which exist", func() {
			BeforeEach(func() {
				configJson := fmt.Sprintf(
					`{"username":"root","password":"%s","host":"%s","port":3306,"database":"%s","adapter":"mysql","tables":["people", "not there"]}`,
					MustHaveEnv("MYSQL_PASSWORD"),
					dbJob.getIPOfInstance(),
					databaseName,
				)
				brJob.runOnVMAndSucceed(fmt.Sprintf("echo '%s' > %s", configJson, configPath))
			})

			It("raises an error about the non-existent tables", func() {
				session := brJob.runOnInstance(fmt.Sprintf(
					"/var/vcap/jobs/database-backup-restorer/bin/backup --artifact-file %s --config %s",
					dbDumpPath,
					configPath))

				Expect(session.ExitCode()).NotTo(BeZero())
				Expect(session).To(gbytes.Say("mysqldump: Couldn't find table: \"not there\""))
				Expect(session).To(gbytes.Say(
					"You may need to delete the artifact-file that was created before re-running"))
			})
		})

		Context("and 'tables' are specified in config none of them exist", func() {
			BeforeEach(func() {
				configJson := fmt.Sprintf(
					`{"username":"root","password":"%s","host":"%s","port":3306,"database":"%s","adapter":"mysql","tables":["lizards", "form-shifting-people"]}`,
					MustHaveEnv("MYSQL_PASSWORD"),
					dbJob.getIPOfInstance(),
					databaseName,
				)
				brJob.runOnVMAndSucceed(fmt.Sprintf("echo '%s' > %s", configJson, configPath))
			})

			It("raises an error about the non-existent tables", func() {
				session := brJob.runOnInstance(fmt.Sprintf(
					"/var/vcap/jobs/database-backup-restorer/bin/backup --artifact-file %s --config %s",
					dbDumpPath,
					configPath))

				Expect(session.ExitCode()).NotTo(BeZero())
				Expect(session).To(gbytes.Say("mysqldump: Couldn't find table: \"lizards\""))
				Expect(session).To(gbytes.Say(
					"You may need to delete the artifact-file that was created before re-running"))
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

			configJson := fmt.Sprintf(
				`{"username":"root","password":"%s","host":"%s","port":3306,"database":"%s","adapter":"mysql"}`,
				MustHaveEnv("MYSQL_PASSWORD"),
				dbJob.getIPOfInstance(),
				databaseName,
			)
			brJob.runOnVMAndSucceed(fmt.Sprintf("echo '%s' > %s", configJson, configPath))
		})

		It("fails with a helpful message", func() {
			backupSession := brJob.runOnInstance(fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/backup --artifact-file %s --config %s", dbDumpPath, configPath))
			Expect(backupSession).To(gexec.Exit(1))
			Expect(backupSession).To(gbytes.Say("Version mismatch"))
		})
	})
})
