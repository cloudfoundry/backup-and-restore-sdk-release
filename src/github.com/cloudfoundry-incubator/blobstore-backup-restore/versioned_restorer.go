package blobstore

import "github.com/cloudfoundry-incubator/blobstore-backup-restore/s3"

type VersionedRestorer struct {
	destinationBuckets map[string]s3.VersionedBucket
	sourceArtifact     VersionedArtifact
}

func NewVersionedRestorer(destinationBuckets map[string]s3.VersionedBucket, sourceArtifact VersionedArtifact) VersionedRestorer {
	return VersionedRestorer{destinationBuckets: destinationBuckets, sourceArtifact: sourceArtifact}
}

func (r VersionedRestorer) Run() error {
	bucketSnapshots, err := r.sourceArtifact.Load()
	if err != nil {
		return err
	}

	for identifier, destinationBucket := range r.destinationBuckets {
		bucketSnapshot := bucketSnapshots[identifier]

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
