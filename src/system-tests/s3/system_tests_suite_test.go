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

package s3_test

import (
	"fmt"
	"os"
	"path"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "system-tests"
)

const (
	awsProfileConfigTemplate = `[profile %s]
    role_arn = %s
    credential_source = Environment`
	assumedRoleProfileName = "assumed-role-profile"
)

var assumedRoleARN = MustHaveEnv("AWS_ASSUMED_ROLE_ARN")

func TestSystemTests(t *testing.T) {
	RegisterFailHandler(Fail)
	SetDefaultEventuallyTimeout(15 * time.Minute)
	RunSpecs(t, "S3 System Tests Suite")
}

var _ = BeforeSuite(func() {
	mustCreateAWSConfigFile()
})

func mustCreateAWSConfigFile() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic("Unable to fetch the user home directory path")
	}
	err = os.Mkdir(path.Join(homeDir, ".aws"), 0755)
	if err != nil {
		panic("Unable to create AWS settings directory")
	}

	configFile, err := os.Create(path.Join(homeDir, ".aws", "config"))
	if err != nil {
		panic("Unable to create AWS settings file")
	}
	defer configFile.Close()

	_, err = fmt.Fprintf(configFile, awsProfileConfigTemplate, assumedRoleProfileName, assumedRoleARN)
	if err != nil {
		panic("Unable to write AWS settings file")
	}
}
