package blobstore

type UnversionedRestorer struct{}

func NewUnversionedRestorer(sourceBuckets map[string]UnversionedBucketPair, destinationArtifact UnversionedArtifact) UnversionedRestorer {
	panic("NewUnversionedRestorer not implemented")
	return UnversionedRestorer{}
}

func (b UnversionedRestorer) Run() error {
	panic("Run not implemented for UnversionedRestorer")
	return nil
}
