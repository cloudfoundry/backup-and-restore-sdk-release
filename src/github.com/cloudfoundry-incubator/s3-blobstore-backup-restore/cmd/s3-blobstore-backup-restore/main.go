package main

import (
	"os"

	"github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/config"

	"github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/unversioned"

	"encoding/json"
	"io/ioutil"

	"errors"
	"flag"

	"fmt"

	"time"

	"github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/incremental"
	"github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/versioned"
)

type Runner interface {
	Run() error
}

func main() {
	commandFlags, err := parseFlags()
	if err != nil {
		exitWithError(err.Error())
	}

	rawConfig, err := ioutil.ReadFile(commandFlags.ConfigPath)
	if err != nil {
		exitWithError("Failed to read config: %s", err.Error())
	}

	var runner Runner
	if commandFlags.Versioned {
		var bucketsConfig map[string]config.BucketConfig
		err = json.Unmarshal(rawConfig, &bucketsConfig)
		if err != nil {
			exitWithError("Failed to parse config: %s", err.Error())
		}

		buckets, err := config.BuildVersionedBuckets(bucketsConfig)
		if err != nil {
			exitWithError("Failed to establish session: %s", err.Error())
		}

		artifact := versioned.NewFileArtifact(commandFlags.ArtifactFilePath)

		if commandFlags.IsRestore {
			runner = versioned.NewRestorer(buckets, artifact)
		} else {
			runner = versioned.NewBackuper(buckets, artifact)
		}
	} else {
		var bucketsConfig map[string]config.UnversionedBucketConfig
		err = json.Unmarshal(rawConfig, &bucketsConfig)
		if err != nil {
			exitWithError("Failed to parse config: %s", err.Error())
		}

		switch {
		case commandFlags.IsRestore:
			{
				incrementalArtifact := incremental.NewArtifact(commandFlags.ArtifactFilePath)
				restoreBucketPairs, err := config.BuildRestoreBucketPairs(bucketsConfig, incrementalArtifact)
				if err != nil {
					exitWithError(fmt.Sprintf("Failed to establish session: %s", err.Error()))
				}

				runner = unversioned.NewRestorer(restoreBucketPairs, incrementalArtifact)
			}
		case commandFlags.UnversionedCompleter:
			{
				existingBackupBlobsArtifact := incremental.NewArtifact(commandFlags.ExistingBackupBlobsArtifactFilePath)
				backupArtifact := incremental.NewArtifact(commandFlags.ArtifactFilePath)
				backupsToComplete, err := config.BuildBackupsToComplete(bucketsConfig, backupArtifact, existingBackupBlobsArtifact)
				if err != nil {
					exitWithError(fmt.Sprintf("Failed to deserialise incremental backups to complete: %s", err.Error()))
				}
				runner = incremental.BackupCompleter{
					BackupsToComplete: backupsToComplete,
				}
			}
		default:
			{
				backupsToStart, err := config.BuildBackupsToStart(bucketsConfig)
				if err != nil {
					exitWithError(fmt.Sprintf("Failed to deserialise incremental backups to start: %s", err.Error()))
				}
				backupArtifact := incremental.NewArtifact(commandFlags.ArtifactFilePath)
				existingBackupBlobsArtifact := incremental.NewArtifact(commandFlags.ExistingBackupBlobsArtifactFilePath)
				runner = incremental.NewBackupStarter(backupsToStart, clock{}, backupArtifact, existingBackupBlobsArtifact)
			}
		}
	}

	err = runner.Run()
	if err != nil {
		exitWithError(err.Error())
	}
}

type clock struct {
}

func (c clock) Now() string {
	return time.Now().Format("2006_01_02_15_04_05")
}

func exitWithError(a ...interface{}) {
	fmt.Fprintln(os.Stderr, a...)
	os.Exit(1)
}

func parseFlags() (CommandFlags, error) {
	var configFilePath = flag.String("config", "", "Path to JSON config file")
	var backupAction = flag.Bool("backup", false, "Run blobstore backup")
	var restoreAction = flag.Bool("restore", false, "Run blobstore restore")
	var artifactFilePath = flag.String("artifact-file", "", "Path to the artifact file")
	var existingBackupBlobsArtifactFilePath = flag.String("existing-backup-blobs-artifact", "", "Path to the existing backup blobs artifact file")
	var unversionedRestore = flag.Bool("unversioned", false, "Indicates targeted buckets are unversioned for restore")
	var unversionedBackupStarter = flag.Bool("unversioned-backup-starter", false, "Run backup starter for unversioned buckets")
	var unversionedBackupCompleter = flag.Bool("unversioned-backup-completer", false, "Run backup completer for unversioned buckets")

	flag.Parse()

	if *backupAction && *restoreAction {
		return CommandFlags{}, errors.New("only one of: --backup or --restore can be provided")
	}

	if !*backupAction && !*restoreAction {
		return CommandFlags{}, errors.New("missing --backup or --restore flag")
	}

	if *configFilePath == "" {
		return CommandFlags{}, errors.New("missing --config flag")
	}

	if *artifactFilePath == "" {
		return CommandFlags{}, errors.New("missing --artifact-file flag")
	}

	return CommandFlags{
		ConfigPath:                          *configFilePath,
		IsRestore:                           *restoreAction,
		ArtifactFilePath:                    *artifactFilePath,
		ExistingBackupBlobsArtifactFilePath: *existingBackupBlobsArtifactFilePath,
		Versioned:                           !*unversionedBackupStarter && !*unversionedBackupCompleter && !*unversionedRestore,
		UnversionedCompleter:                *unversionedBackupCompleter,
	}, nil
}

type CommandFlags struct {
	ConfigPath                          string
	IsRestore                           bool
	ArtifactFilePath                    string
	ExistingBackupBlobsArtifactFilePath string
	Versioned                           bool
	UnversionedCompleter                bool
}
