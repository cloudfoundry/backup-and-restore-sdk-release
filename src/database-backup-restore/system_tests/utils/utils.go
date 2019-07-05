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

package utils

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/onsi/gomega/gexec"

	"time"

	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func RunCommand(cmd string, args ...string) *gexec.Session {
	return RunCommandWithStream(GinkgoWriter, GinkgoWriter, cmd, args...)
}

func RunCommandWithStream(stdout, stderr io.Writer, cmd string, args ...string) *gexec.Session {
	cmdParts := strings.Split(cmd, " ")
	commandPath := cmdParts[0]
	combinedArgs := append(cmdParts[1:], args...)
	command := exec.Command(commandPath, combinedArgs...)

	session, err := gexec.Start(command, stdout, stderr)

	Expect(err).ToNot(HaveOccurred())
	Eventually(session).Should(gexec.Exit())
	return session
}

func BoshCommand() string {
	MustHaveEnv("BOSH_ENVIRONMENT")
	MustHaveEnv("BOSH_CA_CERT")
	MustHaveEnv("BOSH_CLIENT")
	MustHaveEnv("BOSH_CLIENT_SECRET")

	return fmt.Sprintf("bosh-cli --non-interactive")
}

func forDeployment(deploymentName string) string {
	return fmt.Sprintf(
		"--deployment=%s",
		deploymentName,
	)
}

func getSSHCommand(instanceName, instanceIndex string) string {
	MustHaveEnv("BOSH_GW_USER")
	MustHaveEnv("BOSH_GW_HOST")
	MustHaveEnv("BOSH_GW_PRIVATE_KEY")

	return fmt.Sprintf("ssh %s", instanceName)
}

func getUploadCommand(localPath, remotePath, instanceName, instanceIndex string) string {
	MustHaveEnv("BOSH_GW_USER")
	MustHaveEnv("BOSH_GW_HOST")
	MustHaveEnv("BOSH_GW_PRIVATE_KEY")

	return fmt.Sprintf(
		"scp %s %s/%s:%s",
		localPath,
		instanceName,
		instanceIndex,
		remotePath,
	)
}

func getDownloadCommand(remotePath, localPath, instanceName, instanceIndex string) string {
	MustHaveEnv("BOSH_GW_USER")
	MustHaveEnv("BOSH_GW_HOST")
	MustHaveEnv("BOSH_GW_PRIVATE_KEY")

	return fmt.Sprintf(
		"scp %s/%s:%s %s",
		instanceName,
		instanceIndex,
		remotePath,
		localPath,
	)
}

func MustHaveEnv(keyname string) string {
	val := os.Getenv(keyname)
	Expect(val).NotTo(BeEmpty(), "Need "+keyname+" for the test")
	return val
}

func join(args ...string) string {
	return strings.Join(args, " ")
}

type jsonOutputFromCli struct {
	Tables []struct {
		Rows []map[string]string
	}
}

func DisambiguationString() string {
	return strconv.FormatInt(time.Now().UnixNano(), 10)
}

func DisambiguationStringOfLength(length int) string {
	str := DisambiguationString()
	return str[len(str)-length:]
}

func EscapeNewLines(txt string) string {
	return strings.Replace(txt, "\n", "\\n", -1)
}
