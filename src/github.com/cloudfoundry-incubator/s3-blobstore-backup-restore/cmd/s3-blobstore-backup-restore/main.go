package main

import (
	"log"
	"time"

	"github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/config"

	"encoding/json"
	"io/ioutil"

	"errors"
	"flag"

	"fmt"

	"github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/incremental"
	"github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/versioned"
)

type CommandFlags struct {
	ConfigPath                          string
	IsRestore                           bool
	ArtifactFilePath                    string
	ExistingBackupBlobsArtifactFilePath string
	Versioned                           bool
	UnversionedCompleter                bool
}

type Runner interface {
	Run() error
}

type clock struct{}

func (c clock) Now() string {
	return time.Now().Format("2006_01_02_15_04_05")
}

func main() {
	flags, err := parseFlags()
	if err != nil {
		exitWithError("Failed to parse flags", err)
	}

	rawConfig, err := ioutil.ReadFile(flags.ConfigPath)
	if err != nil {
		exitWithError("Failed to read config", err)
	}

	var runner Runner
	if flags.Versioned {
		var bucketsConfig map[string]config.BucketConfig
		err = json.Unmarshal(rawConfig, &bucketsConfig)
		if err != nil {
			exitWithError("Failed to parse config", err)
		}

		buckets, err := config.BuildVersionedBuckets(bucketsConfig)
		if err != nil {
			exitWithError("Failed to establish build versioned buckets", err)
		}

		artifact := versioned.NewFileArtifact(flags.ArtifactFilePath)

		if flags.IsRestore {
			runner = versioned.NewRestorer(buckets, artifact)
		} else {
			runner = versioned.NewBackuper(buckets, artifact)
		}
	} else {
		var bucketsConfig map[string]config.UnversionedBucketConfig
		err = json.Unmarshal(rawConfig, &bucketsConfig)
		if err != nil {
			exitWithError("Failed to parse config", err)
		}

		switch {
		case flags.IsRestore:
			{
				backupArtifact := incremental.NewArtifact(flags.ArtifactFilePath)
				restoreBucketPairs, err := config.BuildRestoreBucketPairs(bucketsConfig, backupArtifact)
				if err != nil {
					exitWithError("Failed to build restore bucket pairs", err)
				}

				runner = incremental.NewRestorer(restoreBucketPairs, backupArtifact)
			}
		case flags.UnversionedCompleter:
			{
				existingBackupBlobsArtifact := incremental.NewArtifact(flags.ExistingBackupBlobsArtifactFilePath)
				backupArtifact := incremental.NewArtifact(flags.ArtifactFilePath)
				backupsToComplete, err := config.BuildBackupsToComplete(bucketsConfig, backupArtifact, existingBackupBlobsArtifact)
				if err != nil {
					exitWithError("Failed to build backups to complete", err)
				}
				runner = incremental.BackupCompleter{
					BackupsToComplete: backupsToComplete,
				}
			}
		default:
			{
				backupsToStart, err := config.BuildBackupsToStart(bucketsConfig)
				if err != nil {
					exitWithError("Failed to build backups to start", err)
				}
				backupArtifact := incremental.NewArtifact(flags.ArtifactFilePath)
				existingBackupBlobsArtifact := incremental.NewArtifact(flags.ExistingBackupBlobsArtifactFilePath)
				runner = incremental.NewBackupStarter(backupsToStart, clock{}, backupArtifact, existingBackupBlobsArtifact)
			}
		}
	}

	err = runner.Run()
	if err != nil {
		exitWithError("Failed to run", err)
	}
}

func exitWithError(context string, err error) {
	log.Fatal(fmt.Sprintf("%s: %s", context, err))
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

	if *unversionedBackupCompleter && *unversionedBackupStarter {
		return CommandFlags{}, errors.New("at most one of: --unversioned-backup-starter or --unversioned-backup-completer can be provided")
	}

	if (*unversionedBackupCompleter || *unversionedBackupStarter) && *existingBackupBlobsArtifactFilePath == "" {
		return CommandFlags{}, errors.New("missing --existing-backup-blobs-artifact")
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
