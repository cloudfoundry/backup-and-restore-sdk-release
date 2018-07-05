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

	"fmt"
	"io"
	"time"
)

type JobInstance struct {
	Deployment          string
	Name                string
	Index               string
	CommandOutputWriter io.Writer
}

func (i *JobInstance) Run(command string) (*gexec.Session, error) {
	return i.runBosh(
		"ssh",
		"--gw-user="+MustHaveEnv("BOSH_GW_USER"),
		"--gw-host="+MustHaveEnv("BOSH_GW_HOST"),
		"--gw-private-key="+MustHaveEnv("BOSH_GW_PRIVATE_KEY"),
		fmt.Sprintf("%s/%s", i.Name, i.Index),
		command,
	)
}

func (i *JobInstance) RunSuccessfully(command string) error {
	session, err := i.Run(command)
	if err != nil {
		return err
	}
	if session.ExitCode() != 0 {
		return fmt.Errorf("bosh command exited non-zero with code %q\n%s", session.ExitCode(), session.Err.Contents())
	}

	return nil
}

func (i *JobInstance) Download(remotePath, localPath string) (*gexec.Session, error) {
	return i.runBosh(
		"scp",
		"--gw-user="+MustHaveEnv("BOSH_GW_USER"),
		"--gw-host="+MustHaveEnv("BOSH_GW_HOST"),
		"--gw-private-key="+MustHaveEnv("BOSH_GW_PRIVATE_KEY"),
		fmt.Sprintf("%s/%s:%s", i.Name, i.Index, remotePath),
		localPath,
	)
}

func (i *JobInstance) Upload(localPath, remotePath string) (*gexec.Session, error) {
	return i.runBosh(
		"scp",
		"--gw-user="+MustHaveEnv("BOSH_GW_USER"),
		"--gw-host="+MustHaveEnv("BOSH_GW_HOST"),
		"--gw-private-key="+MustHaveEnv("BOSH_GW_PRIVATE_KEY"),
		localPath,
		fmt.Sprintf("%s/%s:%s", i.Name, i.Index, remotePath),
	)
}

func (i *JobInstance) runBosh(args ...string) (*gexec.Session, error) {
	combinedArgs := append([]string{
		"--non-interactive",
		"--environment=" + MustHaveEnv("BOSH_ENVIRONMENT"),
		"--deployment=" + i.Deployment,
		"--ca-cert=" + MustHaveEnv("BOSH_CA_CERT"),
		"--client=" + MustHaveEnv("BOSH_CLIENT"),
		"--client-secret=" + MustHaveEnv("BOSH_CLIENT_SECRET"),
	}, args...)
	command := exec.Command("bosh-cli", combinedArgs...)

	session, err := gexec.Start(command, i.CommandOutputWriter, i.CommandOutputWriter)
	if err != nil {
		return session, err
	}

	return waitOnSessionOrError(session, 15*time.Minute)
}

func waitOnSessionOrError(session *gexec.Session, duration time.Duration) (*gexec.Session, error) {
	select {
	case <-time.After(duration):
		return session, fmt.Errorf("command '%s' with args '%s' did not exit after %s", session.Command.Path, session.Command.Args, duration)
	case <-session.Exited:
		return session, nil
	}
}
