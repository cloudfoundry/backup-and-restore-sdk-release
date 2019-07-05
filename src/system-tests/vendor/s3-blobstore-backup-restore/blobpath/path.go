package blobpath

import (
	"strings"
)

const Delimiter = "/"

func Join(prefix, suffix string) string {
	return prefix + Delimiter + suffix
}

func TrimPrefix(path, prefix string) string {
	return strings.TrimPrefix(path, prefix+Delimiter)
}

func TrimTrailingDelimiter(path string) string {
	return strings.TrimSuffix(path, Delimiter)
}
