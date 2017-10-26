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

package main

import (
	"log"

	"github.com/cloudfoundry-incubator/database-backup-restore/config"
	"github.com/cloudfoundry-incubator/database-backup-restore/database"
	"github.com/cloudfoundry-incubator/database-backup-restore/postgres"
)

func main() {
	flags, err := config.ParseFlags()
	if err != nil {
		log.Fatalf("%s\nUsage: database-backup-restorer [--backup|--restore] --config <config-file> "+
			"--artifact-file <artifact-file>\n", err)
	}

	connectionConfig, err := config.ParseAndValidateConnectionConfig(flags.ConfigPath)
	if err != nil {
		log.Fatalf("%v", err)
	}

	utilitiesConfig := config.GetUtilitiesConfigFromEnv()

	interactor, err := makeInteractor(flags.IsRestore, utilitiesConfig, connectionConfig)
	if err != nil {
		log.Fatalf("%v", err)
	}

	err = interactor.Action(flags.ArtifactFilePath)
	if err != nil {
		log.Fatalf(
			"You may need to delete the artifact-file that was created before re-running.\n%s\n", err)
	}
}

func makeInteractor(isRestoreAction bool, utilitiesConfig config.UtilitiesConfig,
	config config.ConnectionConfig) (database.Interactor, error) {

	postgresServerVersionDetector := postgres.NewServerVersionDetector(utilitiesConfig.Postgres96.Client)
	return database.NewInteractorFactory(
		utilitiesConfig,
		postgresServerVersionDetector).Make(actionLabel(isRestoreAction), config)
}

func actionLabel(isRestoreAction bool) database.Action {
	var action database.Action
	if isRestoreAction {
		action = "restore"
	} else {
		action = "backup"
	}
	return action
}
