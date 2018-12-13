package gcs

import (
	"strings"
)

const (
	backupCompleteIdentifier = "backup_complete"
	blobNameDelimiter        = "/"
)

type Blob struct {
	name string `json:"name"`
}

func NewBlob(name string) Blob {
	return Blob{
		name: name,
	}
}

func NewBackupCompleteBlob(prefix string) Blob {
	name := prefix + blobNameDelimiter + backupCompleteIdentifier
	if prefix == "" {
		name = backupCompleteIdentifier
	}

	return Blob{
		name: name,
	}
}

func (b Blob) Name() string {
	return b.name
}

func (b Blob) Resource() string {
	parts := strings.Split(b.name, blobNameDelimiter)
	return parts[len(parts)-1]
}

func (b Blob) IsBackupComplete() bool {
	return b.Resource() == backupCompleteIdentifier
}
