package blobstore

import (
	"github.com/cloudfoundry-incubator/blobstore-backup-restore/s3"
)

//go:generate counterfeiter -o fakes/fake_versioned_bucket.go . VersionedBucket
type VersionedBucket interface {
	Name() string
	RegionName() string
	Versions() ([]s3.Version, error)
	CopyVersions(regionName, bucketName string, versions []BlobVersion) error
}

type S3VersionedBucket struct {
	s3.S3Bucket
}

func NewS3VersionedBucket(s3Bucket s3.S3Bucket) S3VersionedBucket {
	return S3VersionedBucket{s3Bucket}
}

func (bucket S3VersionedBucket) CopyVersions(sourceBucketRegion, sourceBucketName string, versionsToCopy []BlobVersion) error {
	err := bucket.CheckIfVersioned()
	if err != nil {
		return err
	}

	for _, versionToCopy := range versionsToCopy {
		err = bucket.CopyVersion(
			versionToCopy.BlobKey,
			versionToCopy.Id,
			"",
			sourceBucketName,
			sourceBucketRegion,
		)
		if err != nil {
			return err
		}
	}

	return nil
}
