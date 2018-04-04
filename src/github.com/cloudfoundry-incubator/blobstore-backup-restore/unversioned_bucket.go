package blobstore

import (
	"github.com/cloudfoundry-incubator/blobstore-backup-restore/s3"
)

//go:generate counterfeiter -o fakes/fake_unversioned_bucket.go . UnversionedBucket
type UnversionedBucket interface {
	Name() string
	RegionName() string
	Copy(key, destinationPath, originBucketName, originBucketRegion string) error
	ListFiles() ([]string, error)
}

type S3UnversionedBucket struct {
	s3.S3Bucket
}

func NewS3UnversionedBucket(s3Bucket s3.S3Bucket) S3UnversionedBucket {
	return S3UnversionedBucket{s3Bucket}
}
