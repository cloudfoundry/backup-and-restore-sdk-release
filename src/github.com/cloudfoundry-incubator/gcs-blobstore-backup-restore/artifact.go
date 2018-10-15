package gcs

import (
	"encoding/json"
	"io/ioutil"
)

type Artifact struct {
	path string
}

type BackupBucketAddress struct {
	BucketName string `json:"bucket_name"`
	Path       string `json:"path"`
}

func NewArtifact(path string) Artifact {
	return Artifact{path: path}
}

func (a Artifact) Write(backups map[string]BackupBucketAddress) error {

	filesContents, err := json.Marshal(backups)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(a.path, filesContents, 0644)
}

func (a Artifact) Read() (map[string]BackupBucketAddress, error) {
	fileContents, err := ioutil.ReadFile(a.path)
	if err != nil {
		return nil, err
	}

	var backupBuckets = map[string]BackupBucketAddress{}

	err = json.Unmarshal(fileContents, &backupBuckets)
	if err != nil {
		return nil, err
	}

	return backupBuckets, nil
}
