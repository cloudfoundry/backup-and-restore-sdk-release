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

import "fmt"

func BoshCommand() string {
	return fmt.Sprintf("bosh-cli --non-interactive --environment=%s --ca-cert=%s --client=%s --client-secret=%s",
		MustHaveEnv("BOSH_ENVIRONMENT"),
		MustHaveEnv("BOSH_CA_CERT"),
		MustHaveEnv("BOSH_CLIENT"),
		MustHaveEnv("BOSH_CLIENT_SECRET"),
	)
}

func forDeployment(deploymentName string) string {
	return fmt.Sprintf(
		"--deployment=%s",
		deploymentName,
	)
}

func getSSHCommand(instanceName string) string {
	return fmt.Sprintf(
		"ssh --gw-user=%s --gw-host=%s --gw-private-key=%s %s",
		MustHaveEnv("BOSH_GW_USER"),
		MustHaveEnv("BOSH_GW_HOST"),
		MustHaveEnv("BOSH_GW_PRIVATE_KEY"),
		instanceName,
	)
}

func getDownloadCommand(remotePath, localPath, instanceName, instanceIndex string) string {
	return fmt.Sprintf(
		"scp -r --gw-user=%s --gw-host=%s --gw-private-key=%s %s/%s:%s %s",
		MustHaveEnv("BOSH_GW_USER"),
		MustHaveEnv("BOSH_GW_HOST"),
		MustHaveEnv("BOSH_GW_PRIVATE_KEY"),
		instanceName,
		instanceIndex,
		remotePath,
		localPath,
	)
}
func getUploadCommand(localPath, remotePath, instanceName, instanceIndex string) string {
	return fmt.Sprintf(
		"scp -r --gw-user=%s --gw-host=%s --gw-private-key=%s %s %s/%s:%s",
		MustHaveEnv("BOSH_GW_USER"),
		MustHaveEnv("BOSH_GW_HOST"),
		MustHaveEnv("BOSH_GW_PRIVATE_KEY"),
		localPath,
		instanceName,
		instanceIndex,
		remotePath,
	)
}
