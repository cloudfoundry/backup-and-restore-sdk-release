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

		if bucketSnapshot.BucketName != destinationBucket.Name() {
			return fmt.Errorf(
				"artifact bucket name %q for bucket %q does not match configured bucket name %q",
				bucketSnapshot.BucketName, identifier, destinationBucket.Name(),
			)
		}
		if bucketSnapshot.RegionName != destinationBucket.Region() {
			return fmt.Errorf(
				"artifact bucket region %q for bucket %q does not match configured bucket region %q",
				bucketSnapshot.RegionName, identifier, destinationBucket.Region(),
			)
		}
	}

	for identifier, destinationBucket := range r.destinationBuckets {
		bucketSnapshot := bucketSnapshots[identifier]

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
				destinationBucket.Name(),
				destinationBucket.Region(),
			)

			if err != nil {
				return err
			}
		}
	}

	return nil
}
