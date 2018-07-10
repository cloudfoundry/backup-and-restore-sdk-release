package gcs

import (
	"encoding/json"
	"io/ioutil"
)

type Artifact struct {
	path string
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
