package incremental

import "fmt"

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

	for key, _ := range bucketBackups {
		_, exists := b.bucketPairs[key]
		if !exists {
			return fmt.Errorf(
				"bucket %s is not mentioned in the restore config but is present in the artifact",
				key,
			)
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
