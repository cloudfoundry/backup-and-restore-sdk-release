package incremental

import (
	"errors"
	"fmt"
)

type Restorer struct {
	bucketPairs map[string]RestoreBucketPair
	artifact    Artifact
}

func NewRestorer(bucketPairs map[string]RestoreBucketPair, artifact Artifact) Restorer {
	return Restorer{
		bucketPairs: bucketPairs,
		artifact:    artifact,
	}
}

func (b Restorer) Run() error {
	bucketBackups, err := b.artifact.Load()
	if err != nil {
		return err
	}

	for key := range bucketBackups {
		_, exists := b.bucketPairs[key]
		if !exists {
			return fmt.Errorf(
				"restore config does not mention bucket: %s, but is present in the artifact",
				key,
			)
		}

		backupBlobs, _ := b.bucketPairs[key].backupBucket.ListBlobs(bucketBackups[key].SrcBackupDirectoryPath)
		if missingBlobs := validateArtifact(backupBlobs, bucketBackups[key].Blobs); len(missingBlobs) > 0 {
			return formatError(fmt.Sprintf("found blobs in artifact that are not present in backup directory for bucket %s:", bucketBackups[key].BucketName), missingBlobs)
		}
	}

	for key, pair := range b.bucketPairs {
		bucketBackup, exists := bucketBackups[key]
		if !exists {
			return fmt.Errorf("cannot restore bucket %s, not found in backup artifact", key)
		}

		if len(bucketBackup.Blobs) != 0 {
			err = pair.Restore(bucketBackup)
		}

		if err != nil {
			return err
		}
	}
	return nil
}

func validateArtifact(backupBlobs []Blob, artifactBlobs []string) []string {
	var missingBlobs []string
	for _, artifactBlobPath := range artifactBlobs {
		if !contains(artifactBlobPath, backupBlobs) {
			missingBlobs = append(missingBlobs, artifactBlobPath)
		}
	}

	return missingBlobs
}

func contains(key string, blobs []Blob) bool {
	for _, blob := range blobs {
		if key == blob.Path() {
			return true
		}
	}
	return false
}

func formatError(errorMsg string, blobs []string) error {
	for _, blob := range blobs {
		errorMsg = errorMsg + "\n" + blob
	}
	return errors.New(errorMsg)
}
