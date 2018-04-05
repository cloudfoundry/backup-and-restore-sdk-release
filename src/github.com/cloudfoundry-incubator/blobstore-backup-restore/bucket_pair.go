package blobstore

import "github.com/cloudfoundry-incubator/blobstore-backup-restore/s3"

//go:generate counterfeiter -o fakes/fake_unversioned_bucket_pair.go . UnversionedBucketPair
type UnversionedBucketPair interface {
	Backup(backupLocation string) (BackupBucketAddress, error)
}

type S3BucketPair struct {
	LiveBucket   s3.UnversionedBucket
	BackupBucket s3.UnversionedBucket
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
