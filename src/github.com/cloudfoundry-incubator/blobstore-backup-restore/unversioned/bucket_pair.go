package unversioned

import (
	"fmt"

	"github.com/cloudfoundry-incubator/blobstore-backup-restore/s3"
)

//go:generate counterfeiter -o fakes/fake_bucket_pair.go . BucketPair
type BucketPair interface {
	Backup(backupLocation string) (BackupBucketAddress, error)
	Restore(backupLocation string) error
}

type S3BucketPair struct {
	liveBucket   s3.UnversionedBucket
	backupBucket s3.UnversionedBucket
}

func NewS3BucketPair(liveBucket, backupBucket s3.UnversionedBucket) S3BucketPair {
	return S3BucketPair{liveBucket: liveBucket, backupBucket: backupBucket}
}

func (p S3BucketPair) Backup(backupLocation string) (BackupBucketAddress, error) {
	files, err := p.liveBucket.ListFiles("")
	if err != nil {
		return BackupBucketAddress{}, err
	}
	for _, file := range files {
		err = p.backupBucket.CopyObject(file, "", backupLocation, p.liveBucket.Name(), p.liveBucket.RegionName())
		if err != nil {
			return BackupBucketAddress{}, err
		}
	}

	return BackupBucketAddress{
		BucketName:   p.backupBucket.Name(),
		BucketRegion: p.backupBucket.RegionName(),
		Path:         backupLocation,
		EmptyBackup:  len(files) == 0,
	}, nil
}

func (p S3BucketPair) Restore(backupLocation string) error {
	filesToRestore, err := p.backupBucket.ListFiles(backupLocation)
	if err != nil {
		return fmt.Errorf("cannot list files in %s\n %v", p.backupBucket.Name(), err)
	}

	if len(filesToRestore) == 0 {
		return fmt.Errorf("no files found in %s in bucket %s to restore", backupLocation, p.backupBucket.Name())
	}
	for _, file := range filesToRestore {
		err = p.liveBucket.CopyObject(
			file, backupLocation, "", p.backupBucket.Name(), p.backupBucket.RegionName())
		if err != nil {
			return fmt.Errorf("cannot copy object from %s\n %v", p.backupBucket.Name(), err)
		}
	}
	return nil
}
