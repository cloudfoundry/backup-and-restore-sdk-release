package versioned

import (
	"fmt"

	"github.com/cloudfoundry-incubator/blobstore-backup-restore/s3"
)

type Restorer struct {
	destinationBuckets map[string]s3.VersionedBucket
	sourceArtifact     Artifact
}

func NewRestorer(destinationBuckets map[string]s3.VersionedBucket, sourceArtifact Artifact) Restorer {
	return Restorer{destinationBuckets: destinationBuckets, sourceArtifact: sourceArtifact}
}

func (r Restorer) Run() error {
	bucketSnapshots, err := r.sourceArtifact.Load()
	if err != nil {
		return err
	}

	for identifier, destinationBucket := range r.destinationBuckets {
		bucketSnapshot, exists := bucketSnapshots[identifier]

		if !exists {
			return fmt.Errorf("no entry found in backup artifact for bucket: %s", identifier)
		}

		err := destinationBucket.CheckIfVersioned()
		if err != nil {
			return err
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
