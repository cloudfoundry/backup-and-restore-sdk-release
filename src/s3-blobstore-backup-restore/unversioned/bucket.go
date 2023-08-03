package unversioned

import (
	"s3-blobstore-backup-restore/incremental"
	"s3-blobstore-backup-restore/s3bucket"
)

//counterfeiter:generate -o fakes/fake_bucket.go . Bucket
type Bucket interface {
	incremental.Bucket
	IsVersioned() (bool, error)
}

func NewUnversionedBucket(bucketName, bucketRegion, endpoint string, accessKey s3bucket.AccessKey, useIAMProfile, forcePathStyle bool) (bucket Bucket, e error) {
	return s3bucket.NewBucket(bucketName, bucketRegion, endpoint, accessKey, useIAMProfile, forcePathStyle)
}

func NewUnversionedBucketWithRoleARN(bucketName, bucketRegion, endpoint, roleARN string, accessKey s3bucket.AccessKey, useIAMProfile, forcePathStyle bool) (bucket Bucket, e error) {
	return s3bucket.NewBucketWithRoleARN(bucketName, bucketRegion, endpoint, roleARN, accessKey, useIAMProfile, forcePathStyle)
}
