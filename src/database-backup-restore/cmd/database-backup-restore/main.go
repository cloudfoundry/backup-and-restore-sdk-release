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

	"database-backup-restore/config"
	"database-backup-restore/database"
	"database-backup-restore/mysql"
	"database-backup-restore/postgres"
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

	tempFolderManager, err := config.NewTempFolderManager()
	if err != nil {
		log.Fatalf("%v", err)
	}
	defer tempFolderManager.Cleanup()

	interactor, err := makeInteractor(flags.IsRestore, utilitiesConfig, connectionConfig, tempFolderManager)
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
	connectionConfig config.ConnectionConfig, tempFolderManager config.TempFolderManager) (database.Interactor, error) {

	postgresServerVersionDetector := postgres.NewServerVersionDetector(utilitiesConfig.Postgres96.Client)
	mysqlServerVersionDetector := mysql.NewServerVersionDetector(utilitiesConfig.Mysql57.Client)
	interactorFactory := database.NewInteractorFactory(utilitiesConfig, postgresServerVersionDetector, mysqlServerVersionDetector, tempFolderManager)
	return interactorFactory.Make(actionLabel(isRestoreAction), connectionConfig)
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
