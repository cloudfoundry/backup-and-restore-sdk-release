package blobstore

import "github.com/cloudfoundry-incubator/blobstore-backup-restore/s3"

//go:generate counterfeiter -o fakes/fake_unversioned_bucket_pair.go . UnversionedBucketPair
type UnversionedBucketPair interface {
	Backup(backupLocation string) (BackupBucketAddress, error)
}

type S3BucketPair struct {
	LiveBucket   UnversionedBucket
	BackupBucket UnversionedBucket
}

func NewS3BucketPair(liveBucketName, liveBucketRegion, endpoint string, accessKey s3.S3AccessKey,
	backupBucketName string, backupBucketRegion string) (S3BucketPair, error) {

	liveS3Bucket, err := s3.NewBucket(liveBucketName, liveBucketRegion, endpoint, accessKey)
	if err != nil {
		return S3BucketPair{}, err
	}

	backupS3Bucket, err := s3.NewBucket(backupBucketName, backupBucketRegion, endpoint, accessKey)
	if err != nil {
		return S3BucketPair{}, err
	}

	return S3BucketPair{
		LiveBucket:   NewS3UnversionedBucket(liveS3Bucket),
		BackupBucket: NewS3UnversionedBucket(backupS3Bucket),
	}, nil
}

func (p S3BucketPair) Backup(backupLocation string) (BackupBucketAddress, error) {
	files, err := p.LiveBucket.ListFiles()
	if err != nil {
		return BackupBucketAddress{}, err
	}
	for _, file := range files {
		err = p.BackupBucket.Copy(file, backupLocation, p.LiveBucket.Name(), p.LiveBucket.RegionName())
		if err != nil {
			return BackupBucketAddress{}, err
		}
	}
	return BackupBucketAddress{
		BucketName:   p.BackupBucket.Name(),
		BucketRegion: p.BackupBucket.RegionName(),
		Path:         backupLocation,
	}, nil
}
