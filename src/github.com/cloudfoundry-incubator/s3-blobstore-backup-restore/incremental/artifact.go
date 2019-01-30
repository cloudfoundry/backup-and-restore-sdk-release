package incremental

import (
	"encoding/json"
	"io/ioutil"
)

//go:generate counterfeiter -o fakes/fake_artifact.go . Artifact
type Artifact interface {
	Write(map[string]BucketBackup) error
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
	Blobs               []string `json:"blobs"`
	BackupDirectoryPath string   `json:"backup_directory_path"`
}

func (a artifact) Write(backups map[string]BucketBackup) error {
	filesContents, err := json.Marshal(backups)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(a.path, filesContents, 0644)
}
