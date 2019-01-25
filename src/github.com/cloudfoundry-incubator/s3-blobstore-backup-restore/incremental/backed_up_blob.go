package incremental

import "strings"

type BackedUpBlob struct {
	Path                string
	BackupDirectoryPath string
}

func (b BackedUpBlob) LiveBlobPath() string {
	return strings.TrimPrefix(b.Path, b.BackupDirectoryPath+blobDelimiter)
}
