package gcs

import "fmt"

const backupCompleteIdentifier = "backup_complete"

type Blob struct {
	name string `json:"name"`
}

func NewBlob(name string) Blob {
	return Blob{
		name: name,
	}
}

func NewBackupCompleteBlob(prefix string) Blob {
	return Blob{
		name: fmt.Sprintf("%s/%s", prefix, backupCompleteIdentifier),
	}
}

func (b Blob) Name() string {
	return b.name
}

func (b Blob) IsBackupComplete() bool {
	return b.name == backupCompleteIdentifier
}
