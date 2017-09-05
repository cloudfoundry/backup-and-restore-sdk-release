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
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/pivotal-cf/database-backup-and-restore/database"
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
	var configPath = flag.String("config", "", "Path to JSON config file")
	var backupAction = flag.Bool("backup", false, "Run database backup")
	var restoreAction = flag.Bool("restore", false, "Run database restore")
	var artifactFilePath = flag.String("artifact-file", "", "Path to output file")

	flag.Parse()

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

	configString, err := ioutil.ReadFile(*configPath)
	if err != nil {
		log.Fatalf("Fail reading config file: %s\n", err)
	}

	var config database.Config
	if err := json.Unmarshal(configString, &config); err != nil {
		log.Fatalf("Could not parse config json: %s\n", err)
	}

	if !isSupported(config.Adapter) {
		log.Fatalf("Unsupported adapter %s\n", config.Adapter)
	}

	var interactor database.DBInteractor
	if *restoreAction {
		interactor = getRestorer(config, artifactFilePath)
	} else {
		interactor = getBackuper(config, artifactFilePath)
	}

	cmd := interactor.Action()
	fmt.Println(cmd.Args)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatalf("You may need to delete the artifact-file that was created before re-running.\n%s\n", err)
	}
}

func failAndPrintUsage(message string) {
	log.Fatalf("%s\nUsage: database-backup-restorer [--backup|--restore] --config <config-file> --artifact-file <artifact-file>\n", message)
}

func getRestorer(config database.Config, artifactFilePath *string) database.DBInteractor {
	if config.Adapter == "postgres" {
		return database.NewPostgresRestorer(
			config, *artifactFilePath,
		)
	} else {
		return database.NewMysqlRestorer(
			config, *artifactFilePath,
		)
	}
}

func getBackuper(config database.Config, artifactFilePath *string) database.DBInteractor {
	if config.Adapter == "postgres" {
		return database.NewPostgresBackuper(
			config, *artifactFilePath,
		)
	} else {
		return database.NewMysqlBackuper(
			config, *artifactFilePath,
		)
	}
}
