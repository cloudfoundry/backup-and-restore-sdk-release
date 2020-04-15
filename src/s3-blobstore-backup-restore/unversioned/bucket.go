package unversioned

import (
	"s3-blobstore-backup-restore/incremental"
	"s3-blobstore-backup-restore/s3bucket"
)

//go:generate counterfeiter -o fakes/fake_bucket.go . Bucket
type Bucket interface {
	incremental.Bucket
	IsVersioned() (bool, error)
}

func NewUnversionedBucket(bucketName, bucketRegion, endpoint string, accessKey s3bucket.AccessKey, useIAMProfile, usePathStyle bool) (bucket Bucket, e error) {
	return s3bucket.NewBucket(bucketName, bucketRegion, endpoint, accessKey, useIAMProfile, s3bucket.UsePathStyleDuringTheRefactor)
}
