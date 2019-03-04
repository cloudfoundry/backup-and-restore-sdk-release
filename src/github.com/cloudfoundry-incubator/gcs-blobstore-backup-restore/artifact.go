package gcs

import (
	"encoding/json"
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
	Path         string `json:"path"`
	SameBucketAs string `json:",omitempty"`
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
		return nil, err
	}

	var backupBuckets = map[string]BucketBackup{}

	err = json.Unmarshal(fileContents, &backupBuckets)
	if err != nil {
		return nil, err
	}

	return backupBuckets, nil
}
