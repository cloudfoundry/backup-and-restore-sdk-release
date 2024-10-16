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

package integration_tests

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/pivotal-cf-experimental/go-binmock"
)

func TestDatabaseBackupAndRestore(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "DatabaseBackupAndRestore Suite")
}

var compiledSDKPath string
var envVars map[string]string

var fakePgDump13 *binmock.Mock
var fakePgDump15 *binmock.Mock
var fakePgRestore13 *binmock.Mock
var fakePgRestore15 *binmock.Mock
var fakePgClient *binmock.Mock
var fakeMysqlClient57 *binmock.Mock
var fakeMysqlDump57 *binmock.Mock
var fakeMysqlClient80 *binmock.Mock
var fakeMysqlDump80 *binmock.Mock
var fakeMariaDBClient *binmock.Mock
var fakeMariaDBDump *binmock.Mock

var _ = BeforeSuite(func() {
	SetDefaultEventuallyTimeout(5 * time.Second)

	var err error
	compiledSDKPath, err = gexec.Build(
		"database-backup-restore/cmd/database-backup-restore")
	Expect(err).NotTo(HaveOccurred())

	fakePgClient = binmock.NewBinMock(Fail)
	fakePgDump13 = binmock.NewBinMock(Fail)
	fakePgDump15 = binmock.NewBinMock(Fail)
	fakePgRestore13 = binmock.NewBinMock(Fail)
	fakePgRestore15 = binmock.NewBinMock(Fail)
	fakeMysqlDump57 = binmock.NewBinMock(Fail)
	fakeMysqlClient57 = binmock.NewBinMock(Fail)
	fakeMysqlDump80 = binmock.NewBinMock(Fail)
	fakeMysqlClient80 = binmock.NewBinMock(Fail)
	fakeMariaDBClient = binmock.NewBinMock(Fail)
	fakeMariaDBDump = binmock.NewBinMock(Fail)
})

var _ = BeforeEach(func() {
	envVars = map[string]string{
		"PG_CLIENT_PATH":        "non-existent",
		"PG_DUMP_13_PATH":       "non-existent",
		"PG_DUMP_15_PATH":       "non-existent",
		"PG_RESTORE_13_PATH":    "non-existent",
		"PG_RESTORE_15_PATH":    "non-existent",
		"MARIADB_CLIENT_PATH":   "non-existent",
		"MARIADB_DUMP_PATH":     "non-existent",
		"MYSQL_CLIENT_8_0_PATH": "non-existent",
		"MYSQL_DUMP_8_0_PATH":   "non-existent",
		"MYSQL_CLIENT_5_7_PATH": "non-existent",
		"MYSQL_DUMP_5_7_PATH":   "non-existent",
	}
})

var _ = AfterSuite(gexec.CleanupBuildArtifacts)
