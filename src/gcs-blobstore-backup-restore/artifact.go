package gcs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

//go:generate counterfeiter -o fakes/fake_artifact.go . BackupArtifact
type BackupArtifact interface {
	Write(backups map[string]BucketBackup) error
	Read() (map[string]BucketBackup, error)
}

type Artifact struct {
	path string
}

type BucketBackup struct {
	BucketName   string `json:"bucket_name"`
	Bucket       Bucket `json:"-"`
	Path         string `json:"path"`
	SameBucketAs string `json:"same_bucket_as,omitempty"`
}

func NewArtifact(path string) Artifact {
	return Artifact{path: path}
}

func (a Artifact) Write(backups map[string]BucketBackup) error {

	filesContents, err := json.Marshal(backups)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(a.path, filesContents, 0644)
}

func (a Artifact) Read() (map[string]BucketBackup, error) {
	fileContents, err := ioutil.ReadFile(a.path)
	if err != nil {
		return nil, fmt.Errorf("Failed to read artifact file %s: %s", a.path, err.Error())
	}

	var backupBuckets = map[string]BucketBackup{}

	err = json.Unmarshal(fileContents, &backupBuckets)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshall artifact %s: %s", a.path, err.Error())
	}

	return backupBuckets, nil
}
