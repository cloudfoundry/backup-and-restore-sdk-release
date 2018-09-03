package unversioned

import (
	"fmt"

	"strings"

	"errors"

	"github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/execution"
	"github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/s3"
)

//go:generate counterfeiter -o fakes/fake_bucket_pair.go . BucketPair
type BucketPair interface {
	CheckValidity() error
	Backup(backupLocation string) (BackupBucketAddress, error)
	Restore(backupLocation string) error
	LiveBucketName() string
	BackupBucketName() string
}

type S3BucketPair struct {
	liveBucket        s3.UnversionedBucket
	backupBucket      s3.UnversionedBucket
	executionStrategy execution.Strategy
}

func NewS3BucketPair(liveBucket, backupBucket s3.UnversionedBucket, executionStrategy execution.Strategy) S3BucketPair {
	return S3BucketPair{
		liveBucket:        liveBucket,
		backupBucket:      backupBucket,
		executionStrategy: executionStrategy,
	}
}

func (p S3BucketPair) Backup(backupLocation string) (BackupBucketAddress, error) {
	files, err := p.liveBucket.ListFiles("")
	if err != nil {
		return BackupBucketAddress{}, err
	}

	errs := p.executionStrategy.Run(files, func(file string) error {
		return p.backupBucket.CopyObject(file, "", backupLocation, p.liveBucket.Name(), p.liveBucket.RegionName())
	})

	if len(errs) != 0 {
		return BackupBucketAddress{}, formatErrors(
			fmt.Sprintf("failed to backup bucket %s", p.liveBucket.Name()),
			errs,
		)
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

	errs := p.executionStrategy.Run(filesToRestore, func(file string) error {
		return p.liveBucket.CopyObject(file, backupLocation, "", p.backupBucket.Name(), p.backupBucket.RegionName())
	})

	if len(errs) != 0 {
		return formatErrors(
			fmt.Sprintf("failed to restore bucket %s", p.liveBucket.Name()),
			errs,
		)
	}

	return nil
}

func (p S3BucketPair) CheckValidity() error {
	if p.liveBucket.Name() == p.backupBucket.Name() {
		return errors.New("live bucket and backup bucket cannot be the same")
	}

	return nil
}

func formatErrors(contextString string, errors []error) error {
	errorStrings := make([]string, len(errors))
	for i, err := range errors {
		errorStrings[i] = err.Error()
	}
	return fmt.Errorf("%s: %s", contextString, strings.Join(errorStrings, "\n"))
}

func (p S3BucketPair) LiveBucketName() string {
	return p.liveBucket.Name()
}

func (p S3BucketPair) BackupBucketName() string {
	return p.backupBucket.Name()
}
