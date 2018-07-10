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

package backup_and_restore_sdk_release_system_tests

import (
	"os/exec"

	"github.com/onsi/gomega/gexec"
	. "github.com/onsi/gomega/gexec"

	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type JobInstance struct {
	Deployment string
	Name       string
	Index      string
}

func (i *JobInstance) Run(command string) *gexec.Session {
	return i.runBosh(
		"ssh",
		"--gw-user="+MustHaveEnv("BOSH_GW_USER"),
		"--gw-host="+MustHaveEnv("BOSH_GW_HOST"),
		"--gw-private-key="+MustHaveEnv("BOSH_GW_PRIVATE_KEY"),
		fmt.Sprintf("%s/%s", i.Name, i.Index),
		command,
	)
}

func (i *JobInstance) RunSuccessfully(command string) {
	session := i.Run(command)
	Expect(session).To(Exit(0), string(session.Err.Contents()))
}

func (i *JobInstance) Download(remotePath, localPath string) *gexec.Session {
	return i.runBosh(
		"scp",
		"--gw-user="+MustHaveEnv("BOSH_GW_USER"),
		"--gw-host="+MustHaveEnv("BOSH_GW_HOST"),
		"--gw-private-key="+MustHaveEnv("BOSH_GW_PRIVATE_KEY"),
		fmt.Sprintf("%s/%s:%s", i.Name, i.Index, remotePath),
		localPath,
	)
}

func (i *JobInstance) Upload(localPath, remotePath string) *gexec.Session {
	return i.runBosh(
		"scp",
		"--gw-user="+MustHaveEnv("BOSH_GW_USER"),
		"--gw-host="+MustHaveEnv("BOSH_GW_HOST"),
		"--gw-private-key="+MustHaveEnv("BOSH_GW_PRIVATE_KEY"),
		localPath,
		fmt.Sprintf("%s/%s:%s", i.Name, i.Index, remotePath),
	)
}

func (i *JobInstance) runBosh(args ...string) *gexec.Session {
	combinedArgs := append([]string{
		"--non-interactive",
		"--environment=" + MustHaveEnv("BOSH_ENVIRONMENT"),
		"--deployment=" + i.Deployment,
		"--ca-cert=" + MustHaveEnv("BOSH_CA_CERT"),
		"--client=" + MustHaveEnv("BOSH_CLIENT"),
		"--client-secret=" + MustHaveEnv("BOSH_CLIENT_SECRET"),
	}, args...)
	command := exec.Command("bosh-cli", combinedArgs...)

	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())

	Eventually(session).Should(Exit())

	return session
}
