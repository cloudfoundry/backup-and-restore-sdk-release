package config

import (
	"errors"
	"flag"
)

type CommandFlags struct {
	ConfigPath       string
	IsRestore        bool
	ArtifactFilePath string
}

func ParseFlags() (CommandFlags, error) {
	var configPath = flag.String("config", "", "Path to JSON config file")
	var backupAction = flag.Bool("backup", false, "Run database backup")
	var restoreAction = flag.Bool("restore", false, "Run database restore")
	var artifactFilePath = flag.String("artifact-file", "", "Path to output file")
	flag.Parse()
	if *backupAction && *restoreAction {
		return CommandFlags{}, errors.New("Only one of: --backup or --restore can be provided")
	}
	if !*backupAction && !*restoreAction {
		return CommandFlags{}, errors.New("Missing --backup or --restore flag")
	}
	if *configPath == "" {
		return CommandFlags{}, errors.New("Missing --config flag")
	}
	if *artifactFilePath == "" {
		return CommandFlags{}, errors.New("Missing --artifact-file flag")
	}

	return CommandFlags{
		ConfigPath:       *configPath,
		IsRestore:        *restoreAction,
		ArtifactFilePath: *artifactFilePath,
	}, nil
}
