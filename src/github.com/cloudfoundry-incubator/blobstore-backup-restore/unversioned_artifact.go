package blobstore

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

//go:generate counterfeiter -o fakes/fake_unversioned_artifact.go . UnversionedArtifact
type UnversionedArtifact interface {
	Save(backup map[string]BackupBucketAddress) error
	Load() (map[string]BackupBucketAddress, error)
}

type BackupBucketAddress struct {
	BucketName   string `json:"bucket_name"`
	BucketRegion string `json:"bucket_region"`
	Path         string `json:"path"`
}

type UnversionedFileArtifact struct {
	filePath string
}

func NewUnversionedFileArtifact(filePath string) UnversionedFileArtifact {
	return UnversionedFileArtifact{filePath: filePath}
}

func (a UnversionedFileArtifact) Save(backup map[string]BackupBucketAddress) error {
	marshalledBackup, err := json.MarshalIndent(backup, "", "  ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(a.filePath, marshalledBackup, 0666)
	if err != nil {
		return fmt.Errorf("could not write backup file: %s", err.Error())
	}

	return nil
}

func (a UnversionedFileArtifact) Load() (map[string]BackupBucketAddress, error) {
	bytes, err := ioutil.ReadFile(a.filePath)
	if err != nil {
		return nil, fmt.Errorf("could not read backup file: %s", err.Error())
	}

	var backup map[string]BackupBucketAddress
	err = json.Unmarshal(bytes, &backup)
	if err != nil {
		return nil, fmt.Errorf("backup file has an invalid format: %s", err.Error())
	}

	return backup, nil
}
