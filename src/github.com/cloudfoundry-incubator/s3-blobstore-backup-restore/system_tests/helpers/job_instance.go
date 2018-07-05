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

package helpers

import (
	"os/exec"

	. "github.com/cloudfoundry-incubator/system-test-helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/onsi/gomega/gexec"

	"fmt"
)

type JobInstance struct {
	Environment string
	Deployment  string
	Name        string
	Index       string
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

func (i *JobInstance) RunAndSucceed(command string) *gexec.Session {
	session := i.Run(command)
	Expect(session).To(gexec.Exit(0), string(session.Err.Contents()))

	return session
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
		"--environment=" + i.Environment,
		"--deployment=" + i.Deployment,
		"--ca-cert=" + MustHaveEnv("BOSH_CA_CERT"),
		"--client=" + MustHaveEnv("BOSH_CLIENT"),
		"--client-secret=" + MustHaveEnv("BOSH_CLIENT_SECRET"),
	}, args...)
	command := exec.Command("bosh-cli", combinedArgs...)

	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ToNot(HaveOccurred())

	Eventually(session).Should(gexec.Exit())

	return session
}
