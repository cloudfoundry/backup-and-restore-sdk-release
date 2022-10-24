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
	"encoding/json"
	"fmt"
	"os"

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

func (jobInstance *JobInstance) RunPostgresSqlCommand(command, database, user, postgresPackage string) *gexec.Session {
	return jobInstance.RunOnVMAndSucceed(
		fmt.Sprintf(`/var/vcap/packages/%s/bin/psql -U "%s" "%s" --command="%s"`, postgresPackage, user, database, command),
	)
}

func (jobInstance *JobInstance) RunMysqlSqlCommand(command string) *gexec.Session {
	return jobInstance.RunOnVMAndSucceed(
		fmt.Sprintf(`echo -e "%s" | /var/vcap/packages/mariadb/bin/mysql -u root -h localhost --password='%s'`, command, MustHaveEnv("MYSQL_PASSWORD")),
	)
}

func (jobInstance *JobInstance) RunMysqlSqlCommandOnDatabase(database, command string) *gexec.Session {
	return jobInstance.RunOnVMAndSucceed(
		fmt.Sprintf(`echo -e "%s" | /var/vcap/packages/mariadb/bin/mysql -u root -h localhost --password='%s' "%s"`, command, MustHaveEnv("MYSQL_PASSWORD"), database),
	)
}

func (jobInstance *JobInstance) RunOnVMAndSucceed(command string) *gexec.Session {
	session := jobInstance.RunOnInstance(command)
	Expect(session).To(gexec.Exit(0))

	return session
}

func (jobInstance *JobInstance) RunOnInstance(cmd ...string) *gexec.Session {
	if os.Getenv("RUN_TESTS_WITHOUT_BOSH") == "true" {
		return RunCommandWithStream(nil, nil, "bash", "-c", join(cmd...))
	} else {
		return RunCommand(
			join(
				BoshCommand(),
				forDeployment(jobInstance.Deployment),
				getSSHCommand(jobInstance.Instance, jobInstance.InstanceIndex),
			),
			join(cmd...),
		)
	}
}

func (jobInstance *JobInstance) GetIPOfInstance() string {
	session := RunCommand(
		BoshCommand(),
		forDeployment(jobInstance.Deployment),
		"instances",
		"--json",
	)
	outputFromCli := jsonOutputFromCli{}
	contents := session.Out.Contents()
	Expect(json.Unmarshal(contents, &outputFromCli)).To(Succeed())
	for _, instanceData := range outputFromCli.Tables[0].Rows {
		if strings.HasPrefix(instanceData["instance"], jobInstance.Instance+"/") {
			return instanceData["ips"]
		}
	}
	Fail("Cant find instances with name '" + jobInstance.Instance + "' and deployment name '" + jobInstance.Deployment + "'")
	return ""
}

func (jobInstance *JobInstance) downloadFromInstance(remotePath, localPath string) *gexec.Session {
	return RunCommand(
		join(
			BoshCommand(),
			forDeployment(jobInstance.Deployment),
			getDownloadCommand(remotePath, localPath, jobInstance.Instance, jobInstance.InstanceIndex),
		),
	)
}

func (jobInstance *JobInstance) uploadToInstance(localPath, remotePath string) *gexec.Session {
	return RunCommand(
		join(
			BoshCommand(),
			forDeployment(jobInstance.Deployment),
			getUploadCommand(localPath, remotePath, jobInstance.Instance, jobInstance.InstanceIndex),
		),
	)
}
