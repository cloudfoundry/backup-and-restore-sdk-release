// Copyright (C) 2018-Present Pivotal Software, Inc. All rights reserved.
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
	"io"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"strings"
)

type JobInstance struct {
	Deployment    string
	Instance      string
	InstanceIndex string
}

func (jobInstance *JobInstance) RunOnInstanceAndSucceed(command string) *gexec.Session {
	session := jobInstance.runOnInstance(command)
	Expect(session).To(gexec.Exit(0), string(session.Err.Contents()))

	return session
}

func (jobInstance *JobInstance) runOnInstance(cmd ...string) *gexec.Session {
	return runCommand(
		join(
			BoshCommand(),
			forDeployment(jobInstance.Deployment),
			getSSHCommand(jobInstance.Instance),
		),
		join(cmd...),
	)
}

func (jobInstance JobInstance) DownloadFromInstanceAndSucceed(remotePath, localPath string) *gexec.Session {
	session := jobInstance.DownloadFromInstance(remotePath, localPath)
	Expect(session).Should(gexec.Exit(0))

	return session
}

func (jobInstance *JobInstance) DownloadFromInstance(remotePath, localPath string) *gexec.Session {
	return runCommand(
		join(
			BoshCommand(),
			forDeployment(jobInstance.Deployment),
			getDownloadCommand(remotePath, localPath, jobInstance.Instance, jobInstance.InstanceIndex),
		),
	)
}

func runCommand(cmd string, args ...string) *gexec.Session {
	return runCommandWithStream(GinkgoWriter, GinkgoWriter, cmd, args...)
}

func runCommandWithStream(stdout, stderr io.Writer, cmd string, args ...string) *gexec.Session {
	cmdParts := strings.Split(cmd, " ")
	commandPath := cmdParts[0]
	combinedArgs := append(cmdParts[1:], args...)
	command := exec.Command(commandPath, combinedArgs...)

	session, err := gexec.Start(command, stdout, stderr)

	Expect(err).ToNot(HaveOccurred())
	Eventually(session).Should(gexec.Exit())
	return session
}

func join(args ...string) string {
	return strings.Join(args, " ")
}
