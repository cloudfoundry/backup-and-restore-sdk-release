package incremental

import "strings"

const blobDelimiter = "/"

type BackedUpBlob struct {
	Path                string
	BackupDirectoryPath string
}

func (b BackedUpBlob) LiveBlobPath() string {
	return strings.TrimPrefix(b.Path, b.BackupDirectoryPath+blobDelimiter)
}

func joinBlobPath(prefix, suffix string) string {
	return strings.Join([]string{prefix, suffix}, blobDelimiter)
}
