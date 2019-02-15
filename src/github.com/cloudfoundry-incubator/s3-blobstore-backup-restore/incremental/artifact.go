package incremental

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

//go:generate counterfeiter -o fakes/fake_artifact.go . Artifact
type Artifact interface {
	Write(map[string]Backup) error
	Load() (map[string]Backup, error)
}

type artifact struct {
	path string
}

func NewArtifact(path string) Artifact {
	return artifact{
		path: path,
	}
}

type Backup struct {
	BucketName             string   `json:"bucket_name"`
	BucketRegion           string   `json:"bucket_region"`
	Blobs                  []string `json:"blobs"`
	SrcBackupDirectoryPath string   `json:"src_backup_directory_path"`
}

func (a artifact) Write(backups map[string]Backup) error {
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

func (a artifact) Load() (map[string]Backup, error) {
	bytes, err := ioutil.ReadFile(a.path)
	if err != nil {
		return nil, fmt.Errorf("could not read backup file: %s", err.Error())
	}

	var bucketBackups map[string]Backup
	err = json.Unmarshal(bytes, &bucketBackups)
	if err != nil {
		return nil, fmt.Errorf("backup file has an invalid format: %s", err.Error())
	}

	return bucketBackups, nil
}
