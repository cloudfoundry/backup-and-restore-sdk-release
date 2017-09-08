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
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"

	"github.com/cloudfoundry-incubator/database-backup-and-restore/database"
)

var supportedAdapters = []string{"postgres", "mysql"}

func isSupported(adapter string) bool {
	for _, el := range supportedAdapters {
		if el == adapter {
			return true
		}
	}
	return false
}

func main() {
	configPath, backupAction, restoreAction, artifactFilePath := parseFlags()
	validateFlags(backupAction, restoreAction, configPath, artifactFilePath)

	config := readAndValidateConfig(configPath)
	utilitiesConfig := database.GetDependencies()

	interactor := makeInteractor(restoreAction, utilitiesConfig, config)

	if err := interactor.Action(*artifactFilePath); err != nil {
		log.Fatalf("You may need to delete the artifact-file that was created before re-running.\n%s\n", err)
	}
}

func makeInteractor(restoreAction *bool, utilitiesConfig database.UtilitiesConfig, config database.ConnectionConfig) database.DBInteractor {
	if *restoreAction {
		return database.NewRestorerFactory(utilitiesConfig).Make(config)
	} else {
		postgresVersionDetector := database.NewPostgresVersionDetector(utilitiesConfig.Postgres_9_6.Client)
		return database.NewBackuperFactory(utilitiesConfig, postgresVersionDetector).Make(config)
	}
}

func parseFlags() (*string, *bool, *bool, *string) {
	var configPath = flag.String("config", "", "Path to JSON config file")
	var backupAction = flag.Bool("backup", false, "Run database backup")
	var restoreAction = flag.Bool("restore", false, "Run database restore")
	var artifactFilePath = flag.String("artifact-file", "", "Path to output file")
	flag.Parse()
	return configPath, backupAction, restoreAction, artifactFilePath
}

func readAndValidateConfig(configPath *string) database.ConnectionConfig {
	configString, err := ioutil.ReadFile(*configPath)
	if err != nil {
		log.Fatalf("Fail reading config file: %s\n", err)
	}
	var config database.ConnectionConfig
	if err := json.Unmarshal(configString, &config); err != nil {
		log.Fatalf("Could not parse config json: %s\n", err)
	}
	if !isSupported(config.Adapter) {
		log.Fatalf("Unsupported adapter %s\n", config.Adapter)
	}
	if config.Tables != nil && len(config.Tables) == 0 {
		log.Fatalf("Tables specified but empty\n")
	}
	return config
}

func validateFlags(backupAction *bool, restoreAction *bool, configPath *string, artifactFilePath *string) {
	if *backupAction && *restoreAction {
		failAndPrintUsage("Only one of: --backup or --restore can be provided")
	}
	if *configPath == "" {
		failAndPrintUsage("Missing --config flag")
	}
	if !*backupAction && !*restoreAction {
		failAndPrintUsage("Missing --backup or --restore flag")
	}
	if *artifactFilePath == "" {
		failAndPrintUsage("Missing --artifact-file flag")
	}
}

func failAndPrintUsage(message string) {
	log.Fatalf("%s\nUsage: database-backup-restorer [--backup|--restore] --config <config-file> --artifact-file <artifact-file>\n", message)
}
