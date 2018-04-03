package blobstore

type VersionedRestorer struct {
	destinationBuckets map[string]VersionedBucket
	sourceArtifact     VersionedArtifact
}

func NewVersionedRestorer(destinationBuckets map[string]VersionedBucket, sourceArtifact VersionedArtifact) VersionedRestorer {
	return VersionedRestorer{destinationBuckets: destinationBuckets, sourceArtifact: sourceArtifact}
}

func (r VersionedRestorer) Run() error {
	bucketSnapshots, err := r.sourceArtifact.Load()
	if err != nil {
		return err
	}

	for identifier, destinationBucket := range r.destinationBuckets {
		bucketSnapshot := bucketSnapshots[identifier]
		err = destinationBucket.CopyVersions(
			bucketSnapshot.RegionName,
			bucketSnapshot.BucketName,
			bucketSnapshot.Versions)
		if err != nil {
			return err
		}
	}

	return nil
}
