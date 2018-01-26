package blobstore

type Restorer struct {
	destinationBuckets map[string]Bucket
	sourceArtifact     Artifact
}

func NewRestorer(destinationBuckets map[string]Bucket, sourceArtifact Artifact) Restorer {
	return Restorer{destinationBuckets: destinationBuckets, sourceArtifact: sourceArtifact}
}

func (r Restorer) Restore() error {
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
