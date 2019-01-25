package incremental

import (
	"fmt"
	"strings"
)

type BackupToComplete struct {
	BackupBucket    Bucket
	BackupDirectory BackupDirectory
	BlobsToCopy     []BackedUpBlob
}

type BackupCompleter struct {
	BackupsToComplete map[string]BackupToComplete
}

func (b BackupCompleter) Run() error {
	for _, backupToComplete := range b.BackupsToComplete {
		for _, blob := range backupToComplete.BlobsToCopy {
			dst := strings.Join([]string{backupToComplete.BackupDirectory.Path, blob.LiveBlobPath()}, blobDelimiter)

			err := backupToComplete.BackupBucket.CopyBlobWithinBucket(blob.Path, dst)
			if err != nil {

				return fmt.Errorf("failed to complete backup: %s", err)
			}
		}

		err := backupToComplete.BackupDirectory.MarkComplete()
		if err != nil {
			return fmt.Errorf("failed to complete backup: %s", err)
		}
	}

	return nil
}
