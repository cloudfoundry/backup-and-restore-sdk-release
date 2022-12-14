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
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

type TestEntry struct {
	arguments       string
	expectedOutput  string
	configGenerator func() (string, error)
}

var _ = Describe("Backup and Restore DB Utility", func() {
	var compiledSDKPath string
	var artifactFile string
	var err error

	BeforeEach(func() {
		compiledSDKPath, err = gexec.Build(
			"database-backup-restore/cmd/database-backup-restore")
		Expect(err).NotTo(HaveOccurred())

		artifactFile = tempFilePath()
	})

	AfterEach(func() {
		os.Remove(artifactFile)
	})

	Context("incorrect usage or invalid config", func() {
		testCases := []TableEntry{
			Entry("two actions provided", TestEntry{
				arguments:      "--backup --restore",
				expectedOutput: "Only one of: --backup or --restore can be provided",
			}),
			Entry("no action provided", TestEntry{
				arguments:      "--artifact-file /foo --config foo",
				expectedOutput: "Missing --backup or --restore flag",
			}),
			Entry("no config is passed", TestEntry{
				arguments:      "--backup --artifact-file /foo",
				expectedOutput: "Missing --config flag",
			}),
			Entry("the config is not accessible", TestEntry{
				arguments:      "--backup --artifact-file /foo --config /foo/bar/bar.json",
				expectedOutput: "no such file",
			}),
			Entry("the artifact-file is not provided", TestEntry{
				arguments:       "--backup --config %s",
				configGenerator: validPgConfig,
				expectedOutput:  "Missing --artifact-file flag",
			}),
			Entry("is not a valid json", TestEntry{
				arguments:       "--backup --artifact-file /foo --config %s",
				configGenerator: invalidConfig,
				expectedOutput:  "Could not parse config json",
			}),
			Entry("unsupported adapter", TestEntry{
				arguments:       "--backup --artifact-file /foo --config %s",
				configGenerator: invalidAdapterConfig,
				expectedOutput:  "Unsupported adapter foo-server",
			}),
			Entry("empty list of tables field", TestEntry{
				arguments:       "--backup --artifact-file /foo --config %s",
				configGenerator: emptyTablesConfig,
				expectedOutput:  "Tables specified but empty",
			}),
			Entry("tls block without ca", TestEntry{
				arguments:       "--backup --artifact-file /foo --config %s",
				configGenerator: tlsBlockWithoutCaConfig,
				expectedOutput:  "TLS block specified without tls.cert.ca",
			}),
			Entry("client cert without client key", TestEntry{
				arguments:       "--backup --artifact-file /foo --config %s",
				configGenerator: missingClientKeyConfig,
				expectedOutput:  "tls.cert.certificate specified but not tls.cert.private_key",
			}),
			Entry("client key without client cert", TestEntry{
				arguments:       "--backup --artifact-file /foo --config %s",
				configGenerator: missingClientCertConfig,
				expectedOutput:  "tls.cert.private_key specified but not tls.cert.certificate",
			}),
		}

		DescribeTable("raises the appropriate error when",
			func(entry TestEntry) {
				if entry.configGenerator != nil {
					configPath, err := entry.configGenerator()
					Expect(err).NotTo(HaveOccurred())
					entry.arguments = fmt.Sprintf(entry.arguments, configPath)
					defer os.Remove(configPath)
				}

				args := strings.Split(entry.arguments, " ")
				cmd := exec.Command(compiledSDKPath, args...)

				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
				Eventually(session).Should(gexec.Exit(1))
				Expect(session.Err).To(gbytes.Say(entry.expectedOutput))
			},
			testCases...,
		)
	})

	Context("missing environment variables", func() {
		requiredEnvVars := []TableEntry{
			Entry("pg_dump_15 path missing", "PG_DUMP_15_PATH"),
			Entry("pg_restore_15 path missing", "PG_RESTORE_15_PATH"),
			Entry("pg_client path missing", "PG_CLIENT_PATH"),
			Entry("pg_client path missing", "PG_CLIENT_PATH"),
			Entry("pg_dump_13 path missing", "PG_DUMP_13_PATH"),
			Entry("pg_restore_13 path missing", "PG_RESTORE_13_PATH"),
			Entry("pg_dump_11 path missing", "PG_DUMP_11_PATH"),
			Entry("pg_restore_11 path missing", "PG_RESTORE_11_PATH"),
			Entry("pg_client path missing", "PG_CLIENT_PATH"),
			Entry("pg_dump_10 path missing", "PG_DUMP_10_PATH"),
			Entry("pg_restore_10 path missing", "PG_RESTORE_10_PATH"),
			Entry("pg_client path missing", "PG_CLIENT_PATH"),
			Entry("pg_dump_9_6 path missing", "PG_DUMP_9_6_PATH"),
			Entry("pg_restore_9_6 path missing", "PG_RESTORE_9_6_PATH"),
			Entry("pg_client path missing", "PG_CLIENT_PATH"),
			Entry("pg_dump_9_4 path missing", "PG_DUMP_9_4_PATH"),
			Entry("pg_restore_9_4 path missing", "PG_RESTORE_9_4_PATH"),
			Entry("mariadb_client path missing", "MARIADB_CLIENT_PATH"),
			Entry("mariadb_dump path missing", "MARIADB_DUMP_PATH"),
			Entry("mariadb_client path missing", "MARIADB_CLIENT_PATH"),
			Entry("mysql_client_5_6 path missing", "MYSQL_CLIENT_5_6_PATH"),
			Entry("mysql_dump_5_6 path missing", "MYSQL_DUMP_5_6_PATH"),
			Entry("mysql_client_5_6 path missing", "MYSQL_CLIENT_5_6_PATH"),
			Entry("mysql_client_5_7 path missing", "MYSQL_CLIENT_5_7_PATH"),
			Entry("mysql_dump_5_7 path missing", "MYSQL_DUMP_5_7_PATH"),
			Entry("mysql_client_5_7 path missing", "MYSQL_CLIENT_5_7_PATH"),
		}

		DescribeTable("raises the appropriate error when",
			func(missingEnvVar string) {
				configPath, err := validPgConfig()
				Expect(err).NotTo(HaveOccurred())
				defer os.Remove(configPath)

				argumentString :=
					fmt.Sprintf("--backup --artifact-file /foo --config %s", configPath)
				args := strings.Split(argumentString, " ")
				cmd := exec.Command(compiledSDKPath, args...)

				for envVar, val := range envVars {
					if envVar != missingEnvVar {
						cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", envVar, val))
					}
				}

				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).ToNot(HaveOccurred())
				Eventually(session).Should(gexec.Exit(1))
				Eventually(session.Err).Should(gbytes.Say(missingEnvVar + " must be set"))
			},
			requiredEnvVars...,
		)
	})
})

func tempFilePath() string {
	tmpfile, err := os.CreateTemp("", "")
	Expect(err).NotTo(HaveOccurred())
	tmpfile.Close()
	return tmpfile.Name()
}

func invalidConfig() (string, error) {
	invalidJsonConfig, err := os.CreateTemp(os.TempDir(), "")
	if err != nil {
		return "", err
	}
	fmt.Fprint(invalidJsonConfig, "foo!")
	return invalidJsonConfig.Name(), nil
}

func invalidAdapterConfig() (string, error) {
	invalidAdapterConfig, err := os.CreateTemp(os.TempDir(), "")
	if err != nil {
		return "", err
	}
	fmt.Fprint(invalidAdapterConfig, `{"adapter":"foo-server"}`)
	return invalidAdapterConfig.Name(), nil
}

func emptyTablesConfig() (string, error) {
	validConfig, err := os.CreateTemp(os.TempDir(), "")
	if err != nil {
		return "", err
	}

	fmt.Fprint(validConfig,
		`
			{
			  "username":"testuser",
			  "password":"password",
			  "host":"127.0.0.1",
			  "port":1234,
			  "database":"mycooldb",
			  "adapter":"mysql",
			  "tables": []
			}`,
	)
	return validConfig.Name(), nil
}

func tlsBlockWithoutCaConfig() (string, error) {
	validConfig, err := os.CreateTemp(os.TempDir(), "")
	if err != nil {
		return "", err
	}

	fmt.Fprint(validConfig,
		`
			{
			  "username":"testuser",
			  "password":"password",
			  "host":"127.0.0.1",
			  "port":1234,
			  "database":"mycooldb",
			  "adapter":"mysql",
			  "tls": {	}
			}`,
	)
	return validConfig.Name(), nil
}

func missingClientKeyConfig() (string, error) {
	validConfig, err := os.CreateTemp(os.TempDir(), "")
	if err != nil {
		return "", err
	}

	fmt.Fprint(validConfig,
		`
			{
			  "username":"testuser",
			  "password":"password",
			  "host":"127.0.0.1",
			  "port":1234,
			  "database":"mycooldb",
			  "adapter":"mysql",
			  "tls": {
					"cert": {
						"ca": "ca cert",
						"certificate": "client certificate"
					}
				}
			}`,
	)
	return validConfig.Name(), nil
}

func missingClientCertConfig() (string, error) {
	validConfig, err := os.CreateTemp(os.TempDir(), "")
	if err != nil {
		return "", err
	}

	fmt.Fprint(validConfig,
		`
			{
			  "username":"testuser",
			  "password":"password",
			  "host":"127.0.0.1",
			  "port":1234,
			  "database":"mycooldb",
			  "adapter":"mysql",
			  "tls": {
					"cert": {
						"ca": "ca cert",
						"private_key": "client private key"
					}
				}
			}`,
	)
	return validConfig.Name(), nil
}

func validPgConfig() (string, error) {
	validConfig, err := os.CreateTemp(os.TempDir(), "")
	if err != nil {
		return "", err
	}

	fmt.Fprint(validConfig,
		`
		  {
		    "username":"testuser",
		    "password":"password",
		    "host":"127.0.0.1",
		    "port":1234,
		    "database":"mycooldb",
		    "adapter":"postgres"
		  }`,
	)
	return validConfig.Name(), nil
}

func saveFile(content string) *os.File {
	configFile, err := os.CreateTemp(os.TempDir(), time.Now().String())
	Expect(err).NotTo(HaveOccurred())

	os.WriteFile(configFile.Name(), []byte(content), 0755)

	return configFile
}
