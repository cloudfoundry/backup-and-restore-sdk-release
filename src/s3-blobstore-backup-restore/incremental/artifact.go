package incremental

import (
	"encoding/json"
	"fmt"
	"os"
)

//counterfeiter:generate -o fakes/fake_artifact.go . Artifact
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
	BucketName             string   `json:"bucket_name,omitempty"`
	BucketRegion           string   `json:"bucket_region,omitempty"`
	Blobs                  []string `json:"blobs"`
	SrcBackupDirectoryPath string   `json:"src_backup_directory_path"`
	DstBackupDirectoryPath string   `json:"dst_backup_directory_path,omitempty"`
	SameBucketAs           string   `json:",omitempty"`
}

func (a artifact) Write(backups map[string]Backup) error {
	filesContents, err := json.Marshal(backups)
	if err != nil {
		return err
	}

	err = os.WriteFile(a.path, filesContents, 0644)
	if err != nil {
		return fmt.Errorf("could not write backup file: %s", err.Error())
	}

	return nil
}

func (a artifact) Load() (map[string]Backup, error) {
	bytes, err := os.ReadFile(a.path)
	if err != nil {
		return nil, fmt.Errorf("could not read backup file: %s", err.Error())
	}

	var backups map[string]Backup
	err = json.Unmarshal(bytes, &backups)
	if err != nil {
		return nil, fmt.Errorf("backup file has an invalid format: %s", err.Error())
	}

	return backups, nil
}
