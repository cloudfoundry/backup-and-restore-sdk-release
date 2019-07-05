package versioned

import (
	"fmt"
)

type Restorer struct {
	destinationBuckets map[string]Bucket
	sourceArtifact     Artifact
}

func NewRestorer(destinationBuckets map[string]Bucket, sourceArtifact Artifact) Restorer {
	return Restorer{destinationBuckets: destinationBuckets, sourceArtifact: sourceArtifact}
}

func (r Restorer) Run() error {
	bucketSnapshots, err := r.sourceArtifact.Load()
	if err != nil {
		return err
	}

	for identifier := range bucketSnapshots {
		_, exists := r.destinationBuckets[identifier]

		if !exists {
			return fmt.Errorf("no entry found in restore config for bucket: %s", identifier)
		}
	}

	for identifier, destinationBucket := range r.destinationBuckets {
		bucketSnapshot, exists := bucketSnapshots[identifier]

		if !exists {
			return fmt.Errorf("no entry found in backup artifact for bucket: %s", identifier)
		}

		isVersioned, err := destinationBucket.IsVersioned()

		if err != nil {
			return fmt.Errorf("failed to check if %s is versioned: %s", destinationBucket.Name(), err.Error())
		}

		if !isVersioned {
			return fmt.Errorf("bucket %s is not versioned", destinationBucket.Name())
		}

		for _, versionToCopy := range bucketSnapshot.Versions {
			err = destinationBucket.CopyVersion(
				versionToCopy.BlobKey,
				versionToCopy.Id,
				bucketSnapshot.BucketName,
				bucketSnapshot.RegionName,
			)

			if err != nil {
				return err
			}
		}
	}

	return nil
}
