package s3bucket

func (b Bucket) GetBlobSizeImpl(bucketName, bucketRegion, blobKey, versionID string) (int64, error) {
	return b.getBlobSize(bucketName, bucketRegion, blobKey, versionID)
}
