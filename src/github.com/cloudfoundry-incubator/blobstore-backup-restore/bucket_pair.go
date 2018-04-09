package blobstore

import (
	"fmt"

	"github.com/cloudfoundry-incubator/blobstore-backup-restore/s3"
)

//go:generate counterfeiter -o fakes/fake_unversioned_bucket_pair.go . UnversionedBucketPair
type UnversionedBucketPair interface {
	Backup(backupLocation string) (BackupBucketAddress, error)
	Restore(backupLocation string) error
}

type S3BucketPair struct {
	LiveBucket   s3.UnversionedBucket
	BackupBucket s3.UnversionedBucket
}

func (p S3BucketPair) Backup(backupLocation string) (BackupBucketAddress, error) {
	files, err := p.LiveBucket.ListFiles("")
	if err != nil {
		return BackupBucketAddress{}, err
	}
	for _, file := range files {
		err = p.BackupBucket.CopyObject(file, "", backupLocation, p.LiveBucket.Name(), p.LiveBucket.RegionName())
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

func (p S3BucketPair) Restore(backupLocation string) error {
	filesToRestore, err := p.BackupBucket.ListFiles(backupLocation)
	if err != nil {
		return fmt.Errorf("cannot list files in %s\n %v", p.BackupBucket.Name(), err)
	}
	for _, file := range filesToRestore {
		err = p.LiveBucket.CopyObject(file, backupLocation, "", p.BackupBucket.Name(), p.BackupBucket.RegionName())
		if err != nil {
			return fmt.Errorf("cannot copy object from %s\n %v", p.BackupBucket.Name(), err)
		}
	}
	return nil
}
