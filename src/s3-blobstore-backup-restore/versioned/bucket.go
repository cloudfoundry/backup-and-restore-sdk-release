package versioned

import "github.com/cloudfoundry/backup-and-restore-sdk-release/src/s3-blobstore-backup-restore/s3bucket"

//go:generate counterfeiter -o fakes/fake_bucket.go . Bucket
type Bucket interface {
	Name() string
	Region() string
	CopyVersion(blobKey, versionId, originBucketName, originBucketRegion string) error
	ListVersions() ([]s3bucket.Version, error)
	IsVersioned() (bool, error)
}

func NewVersionedBucket(bucketName, bucketRegion, endpoint string, accessKey s3bucket.AccessKey, useIAMProfile, forcePathStyle bool) (bucket Bucket, e error) {
	return s3bucket.NewBucket(bucketName, bucketRegion, endpoint, accessKey, useIAMProfile, forcePathStyle)
}
