package blobstore

import "fmt"

type UnversionedRestorer struct {
	bucketPairs map[string]UnversionedBucketPair
	artifact    UnversionedArtifact
}

func NewUnversionedRestorer(bucketPairs map[string]UnversionedBucketPair, artifact UnversionedArtifact) UnversionedRestorer {
	return UnversionedRestorer{
		bucketPairs: bucketPairs,
		artifact:    artifact,
	}
}

func (b UnversionedRestorer) Run() error {
	backupBucketAddresses, err := b.artifact.Load()
	if err != nil {
		return err
	}

	for key, _ := range backupBucketAddresses {
		_, exists := b.bucketPairs[key]
		if !exists {
			return fmt.Errorf(
				"bucket %s is not mentioned in the restore config but is present in the artifact",
				key,
			)
		}
	}

	for key, pair := range b.bucketPairs {
		address, exists := backupBucketAddresses[key]
		if !exists {
			return fmt.Errorf("cannot restore bucket %s, not found in backup artifact", key)
		}
		err = pair.Restore(address.Path)
		if err != nil {
			return err
		}
	}
	return nil
}
