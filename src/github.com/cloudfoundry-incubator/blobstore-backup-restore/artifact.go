package blobstore

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

//go:generate counterfeiter -o fakes/fake_versioned_artifact.go . VersionedArtifact
type VersionedArtifact interface {
	Save(backup map[string]BucketSnapshot) error
	Load() (map[string]BucketSnapshot, error)
}

//go:generate counterfeiter -o fakes/fake_unversioned_artifact.go . UnversionedArtifact
type UnversionedArtifact interface {
	Save(backup map[string]BackupBucketAddress) error
	Load() (map[string]BackupBucketAddress, error)
}

type VersionedFileArtifact struct {
	filePath string
}

func NewVersionedFileArtifact(filePath string) VersionedFileArtifact {
	return VersionedFileArtifact{filePath: filePath}
}

type UnversionedFileArtifact struct {
	filePath string
}

func NewUnversionedFileArtifact(filePath string) UnversionedFileArtifact {
	return UnversionedFileArtifact{filePath: filePath}
}

func (a VersionedFileArtifact) Save(backup map[string]BucketSnapshot) error {
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

func (a VersionedFileArtifact) Load() (map[string]BucketSnapshot, error) {
	bytes, err := ioutil.ReadFile(a.filePath)
	if err != nil {
		return nil, fmt.Errorf("could not read backup file: %s", err.Error())
	}

	var backup map[string]BucketSnapshot
	err = json.Unmarshal(bytes, &backup)
	if err != nil {
		return nil, fmt.Errorf("backup file has an invalid format: %s", err.Error())
	}

	return backup, nil
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

type BucketSnapshot struct {
	BucketName string        `json:"bucket_name"`
	RegionName string        `json:"region_name"`
	Versions   []BlobVersion `json:"versions"`
}

type BlobVersion struct {
	BlobKey string `json:"blob_key"`
	Id      string `json:"version_id"`
}

type BackupBucketAddress struct {
	BucketName   string `json:"bucket_name"`
	BucketRegion string `json:"bucket_region"`
	Path         string `json:"path"`
}
