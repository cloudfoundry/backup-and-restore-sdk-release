package blobstore

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

//go:generate counterfeiter -o fakes/fake_artifact.go . Artifact
type Artifact interface {
	Save(backup map[string]BucketBackup) error
	Load() (map[string]BucketBackup, error)
}

type FileArtifact struct {
	filePath string
}

func NewFileArtifact(filePath string) FileArtifact {
	return FileArtifact{filePath: filePath}
}

func (a FileArtifact) Save(backup map[string]BucketBackup) error {
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

func (a FileArtifact) Load() (map[string]BucketBackup, error) {
	bytes, err := ioutil.ReadFile(a.filePath)
	if err != nil {
		return nil, fmt.Errorf("could not read backup file: %s", err.Error())
	}

	var backup map[string]BucketBackup
	err = json.Unmarshal(bytes, &backup)
	if err != nil {
		return nil, fmt.Errorf("backup file has an invalid format: %s", err.Error())
	}

	return backup, nil
}

type BucketBackup struct {
	BucketName string          `json:"bucket_name"`
	RegionName string          `json:"region_name"`
	Versions   []LatestVersion `json:"versions"`
}

type LatestVersion struct {
	BlobKey string `json:"blob_key"`
	Id      string `json:"version_id"`
}
