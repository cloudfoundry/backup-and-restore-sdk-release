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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"

	"fmt"

	"os"

	. "github.com/cloudfoundry-incubator/database-backup-restore/system_tests/utils"

	_ "github.com/lib/pq"
)

var _ = Describe("postgres", func() {
	var databaseName string
	var dbDumpPath string
	var configPath string

	BeforeEach(func() {
		disambiguationString := DisambiguationString()
		databaseName = "db" + disambiguationString
		configPath = "/tmp/config" + disambiguationString
		dbDumpPath = "/tmp/artifact" + disambiguationString

		RunSQLCommand("CREATE DATABASE "+databaseName, connection)
		connection.Close()

		connection, proxySession = SuccessfullyConnectToPostgres(
			postgresHostName,
			postgresPassword,
			postgresNonSslUsername,
			postgresPort,
			databaseName,
			os.Getenv("SSH_PROXY_HOST"),
			os.Getenv("SSH_PROXY_USER"),
			os.Getenv("SSH_PROXY_KEY_FILE"),
		)

		RunSQLCommand("CREATE TABLE people (name varchar);", connection)
		RunSQLCommand("INSERT INTO people VALUES ('Old Person');", connection)
		RunSQLCommand("CREATE TABLE places (name varchar);", connection)
		RunSQLCommand("INSERT INTO places VALUES ('Old Place');", connection)
	})

	AfterEach(func() {

		connection.Close()

		connection, proxySession = SuccessfullyConnectToPostgres(
			postgresHostName,
			postgresPassword,
			postgresNonSslUsername,
			postgresPort,
			"postgres",
			os.Getenv("SSH_PROXY_HOST"),
			os.Getenv("SSH_PROXY_USER"),
			os.Getenv("SSH_PROXY_KEY_FILE"),
		)

		RunSQLCommand("DROP DATABASE "+databaseName, connection)
		brJob.RunOnVMAndSucceed(fmt.Sprintf("sudo rm -rf %s %s", configPath, dbDumpPath))
	})

	Context("database dump is successful", func() {
		BeforeEach(func() {
			configJson := fmt.Sprintf(
				`{"username":"test_user","password":"%s","host":"%s","port":5432,
							"database":"%s","adapter":"postgres"}`,
				postgresPassword,
				postgresHostName,
				databaseName,
			)
			brJob.RunOnVMAndSucceed(fmt.Sprintf("echo '%s' > %s", configJson, configPath))
		})

		It("backs up the Postgres database", func() {
			brJob.RunOnVMAndSucceed(
				fmt.Sprintf(`/var/vcap/jobs/database-backup-restorer/bin/backup --config %s --artifact-file %s`,
					configPath, dbDumpPath))
			brJob.RunOnVMAndSucceed(fmt.Sprintf("ls -l %s", dbDumpPath))

			RunSQLCommand("UPDATE people SET NAME = 'New Person';", connection)
			RunSQLCommand("UPDATE places SET NAME = 'New Place';", connection)

			brJob.RunOnVMAndSucceed(
				fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/restore --config %s --artifact-file %s",
					configPath, dbDumpPath))

			Expect(FetchSQLColumn("SELECT name FROM people;", connection)).
				To(ConsistOf("Old Person"))
			Expect(FetchSQLColumn("SELECT name FROM people;", connection)).
				NotTo(ConsistOf("New Person"))
			Expect(FetchSQLColumn("SELECT name FROM places;", connection)).
				To(ConsistOf("Old Place"))
			Expect(FetchSQLColumn("SELECT name FROM places;", connection)).
				NotTo(ConsistOf("New Place"))
		})

	})

	Context("and 'tables' are specified in config", func() {
		BeforeEach(func() {
			configJson := fmt.Sprintf(
				`{"username":"test_user","password":"%s","host":"%s","port":5432,
							"database":"%s","adapter":"postgres", "tables":["people"]}`,
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

			RunSQLCommand("UPDATE people SET NAME = 'New Person';", connection)
			RunSQLCommand("UPDATE places SET NAME = 'New Place';", connection)

			brJob.RunOnVMAndSucceed(fmt.Sprintf("cat %s", dbDumpPath))

			brJob.RunOnVMAndSucceed(fmt.Sprintf("/var/vcap/jobs/database-backup-restorer/bin/restore --artifact-file %s --config %s", dbDumpPath, configPath))

			Expect(FetchSQLColumn("SELECT name FROM people;", connection)).
				To(ConsistOf("Old Person"))
			Expect(FetchSQLColumn("SELECT name FROM people;", connection)).
				NotTo(ConsistOf("New Person"))
			Expect(FetchSQLColumn("SELECT name FROM places;", connection)).
				To(ConsistOf("New Place"))
			Expect(FetchSQLColumn("SELECT name FROM places;", connection)).
				NotTo(ConsistOf("Old Place"))
		})
	})

	Context("and 'tables' are specified in config, with a non-existent table", func() {
		BeforeEach(func() {
			configJson := fmt.Sprintf(
				`{"username":"test_user","password":"%s","host":"%s","port":5432,
							"database":"%s","adapter":"postgres", "tables":["people", "lizards"]}`,
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
