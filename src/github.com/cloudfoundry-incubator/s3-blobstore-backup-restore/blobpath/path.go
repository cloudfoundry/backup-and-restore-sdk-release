package blobpath

import (
	"path/filepath"
	"strings"
)

const Delimiter = "/"

func Join(prefix, suffix string) string {
	return filepath.Join(prefix, suffix)
}

func TrimPrefix(path, prefix string) string {
	return strings.TrimPrefix(path, prefix+Delimiter)
}

func TrimTrailingDelimiter(path string) string {
	return strings.TrimSuffix(path, Delimiter)
}
