package incremental

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

//go:generate counterfeiter -o fakes/fake_artifact.go . Artifact
type Artifact interface {
	Write(map[string]BucketBackup) error
	Load() (map[string]BucketBackup, error)
}

type artifact struct {
	path string
}

func NewArtifact(path string) Artifact {
	return artifact{
		path: path,
	}
}

type BucketBackup struct {
	BucketName          string   `json:"bucket_name"`
	BucketRegion        string   `json:"bucket_region"`
	Blobs               []string `json:"blobs"`
	BackupDirectoryPath string   `json:"backup_directory_path"`
}

func (a artifact) Write(backups map[string]BucketBackup) error {
	filesContents, err := json.Marshal(backups)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(a.path, filesContents, 0644)
	if err != nil {
		return fmt.Errorf("could not write backup file: %s", err.Error())
	}

	return nil
}

func (a artifact) Load() (map[string]BucketBackup, error) {
	bytes, err := ioutil.ReadFile(a.path)
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

func (a artifact) Load() (map[string]BucketBackup, error) {
	content, err := ioutil.ReadFile(a.path)
	if err != nil {
		return nil, err
	}

	bucketBackups := new(map[string]BucketBackup)

	err = json.Unmarshal(content, &bucketBackups)
	if err != nil {
		return nil, err
	}

	return *bucketBackups, nil
}
