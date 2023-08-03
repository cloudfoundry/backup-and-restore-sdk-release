package versioned

import (
	"encoding/json"
	"fmt"
	"os"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate -o fakes/fake_artifact.go . Artifact
type Artifact interface {
	Save(backup map[string]BucketSnapshot) error
	Load() (map[string]BucketSnapshot, error)
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

type FileArtifact struct {
	filePath string
}

func NewFileArtifact(filePath string) FileArtifact {
	return FileArtifact{filePath: filePath}
}

func (a FileArtifact) Save(backup map[string]BucketSnapshot) error {
	marshalledBackup, err := json.MarshalIndent(backup, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile(a.filePath, marshalledBackup, 0666)
	if err != nil {
		return fmt.Errorf("could not write backup file: %s", err.Error())
	}

	return nil
}

func (a FileArtifact) Load() (map[string]BucketSnapshot, error) {
	bytes, err := os.ReadFile(a.filePath)
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
