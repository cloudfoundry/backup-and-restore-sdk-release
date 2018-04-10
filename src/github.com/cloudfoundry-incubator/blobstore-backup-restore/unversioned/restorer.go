package unversioned

import "fmt"

type Restorer struct {
	bucketPairs map[string]BucketPair
	artifact    Artifact
}

func NewRestorer(bucketPairs map[string]BucketPair, artifact Artifact) Restorer {
	return Restorer{
		bucketPairs: bucketPairs,
		artifact:    artifact,
	}
}

func (b Restorer) Run() error {
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
		if !address.EmptyBackup {
			err = pair.Restore(address.Path)
		}
		if err != nil {
			return err
		}
	}
	return nil
}
