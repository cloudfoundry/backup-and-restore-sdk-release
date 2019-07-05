package gcs

import (
	"strings"
)

const (
	blobNameDelimiter = "/"
)

type Blob struct {
	name string `json:"name"`
}

func NewBlob(name string) Blob {
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
