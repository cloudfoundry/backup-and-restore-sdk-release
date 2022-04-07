package main

import (
	"log"
	"time"

	"github.com/cloudfoundry/backup-and-restore-sdk-release/src/s3-blobstore-backup-restore/unversioned"

	"encoding/json"
	"io/ioutil"

	"errors"
	"flag"

	"fmt"

	"github.com/cloudfoundry/backup-and-restore-sdk-release/src/s3-blobstore-backup-restore/incremental"
	"github.com/cloudfoundry/backup-and-restore-sdk-release/src/s3-blobstore-backup-restore/versioned"
)

type CommandFlags struct {
	ConfigPath                          string
	ArtifactFilePath                    string
	ExistingBackupBlobsArtifactFilePath string

	VersionedBackup  *bool
	VersionedRestore *bool

	UnversionedBackupStart    *bool
	UnversionedBackupComplete *bool
	UnversionedRestore        *bool
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
	if *flags.VersionedBackup || *flags.VersionedRestore {
		var bucketsConfig map[string]versioned.BucketConfig
		err = json.Unmarshal(rawConfig, &bucketsConfig)
		if err != nil {
			exitWithError("Failed to parse config", err)
		}

		buckets, err := versioned.BuildVersionedBuckets(bucketsConfig, versioned.NewVersionedBucket)
		if err != nil {
			exitWithError("Failed to establish build versioned buckets", err)
		}

		artifact := versioned.NewFileArtifact(flags.ArtifactFilePath)

		if *flags.VersionedBackup {
			runner = versioned.NewBackuper(buckets, artifact)
		} else {
			runner = versioned.NewRestorer(buckets, artifact)
		}
	} else {
		var bucketsConfig map[string]unversioned.UnversionedBucketConfig
		err = json.Unmarshal(rawConfig, &bucketsConfig)
		if err != nil {
			exitWithError("Failed to parse config", err)
		}

		if *flags.UnversionedBackupStart {
			backupsToStart, err := unversioned.BuildBackupsToStart(bucketsConfig, unversioned.NewUnversionedBucket)
			if err != nil {
				exitWithError("Failed to build backups to start", err)
			}
			backupArtifact := incremental.NewArtifact(flags.ArtifactFilePath)
			existingBackupBlobsArtifact := incremental.NewArtifact(flags.ExistingBackupBlobsArtifactFilePath)
			runner = incremental.NewBackupStarter(backupsToStart, clock{}, backupArtifact, existingBackupBlobsArtifact)
		} else if *flags.UnversionedBackupComplete {
			existingBackupBlobsArtifact := incremental.NewArtifact(flags.ExistingBackupBlobsArtifactFilePath)
			backupsToComplete, err := unversioned.BuildBackupsToComplete(bucketsConfig, existingBackupBlobsArtifact, unversioned.NewUnversionedBucket)
			if err != nil {
				exitWithError("Failed to build backups to complete", err)
			}
			runner = incremental.BackupCompleter{
				BackupsToComplete: backupsToComplete,
			}
		} else {
			backupArtifact := incremental.NewArtifact(flags.ArtifactFilePath)
			restoreBucketPairs, err := unversioned.BuildRestoreBucketPairs(bucketsConfig, backupArtifact, unversioned.NewUnversionedBucket)
			if err != nil {
				exitWithError("Failed to build restore bucket pairs", err)
			}

			runner = incremental.NewRestorer(restoreBucketPairs, backupArtifact)
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
	var (
		configFilePath       = flag.String("config", "", "Path to config file")
		artifactPath         = flag.String("artifact", "", "Path to the artifact file")
		existingArtifactPath = flag.String("existing-artifact", "", "Path to the existing backups artifact file")
		versionedBackup      = flag.Bool("versioned-backup", false, "Run versioned blobstore backup")
		versionedRestore     = flag.Bool("versioned-restore", false, "Run versioned blobstore restore")

		unversionedBackupStart    = flag.Bool("unversioned-backup-start", false, "Run backup starter for unversioned buckets")
		unversionedBackupComplete = flag.Bool("unversioned-backup-complete", false, "Run backup completer for unversioned buckets")
		unversionedRestore        = flag.Bool("unversioned-restore", false, "Run unversioned blobstore restore")
	)

	flag.Parse()

	if *configFilePath == "" {
		return CommandFlags{}, errors.New("missing --config flag")
	}

	var count int
	for _, b := range []*bool{versionedBackup, versionedRestore, unversionedBackupStart, unversionedBackupComplete, unversionedRestore} {
		if *b {
			count++
		}
	}

	if count != 1 {
		return CommandFlags{}, errors.New("exactly one action flag must be provided")
	}

	if (*versionedBackup || *versionedRestore || *unversionedBackupStart || *unversionedRestore) && *artifactPath == "" {
		return CommandFlags{}, errors.New("missing --artifact flag")
	}

	if (*unversionedBackupStart || *unversionedBackupComplete) && *existingArtifactPath == "" {
		return CommandFlags{}, errors.New("missing --existing-artifact flag")
	}

	return CommandFlags{
		ConfigPath:                          *configFilePath,
		ArtifactFilePath:                    *artifactPath,
		ExistingBackupBlobsArtifactFilePath: *existingArtifactPath,

		VersionedBackup:  versionedBackup,
		VersionedRestore: versionedRestore,

		UnversionedBackupStart:    unversionedBackupStart,
		UnversionedBackupComplete: unversionedBackupComplete,
		UnversionedRestore:        unversionedRestore,
	}, nil
}
