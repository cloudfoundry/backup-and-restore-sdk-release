package unversioned

import (
	"errors"
	"fmt"

	"strings"

	"github.com/cloudfoundry-incubator/bosh-backup-and-restore/executor"
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
	executionStrategy executor.ParallelExecutor
}

func NewS3BucketPair(liveBucket, backupBucket s3.UnversionedBucket) S3BucketPair {
	exe := executor.NewParallelExecutor()
	exe.SetMaxInFlight(200)
	return S3BucketPair{
		liveBucket:        liveBucket,
		backupBucket:      backupBucket,
		executionStrategy: exe,
	}
}

func (p S3BucketPair) Backup(backupLocation string) (BackupBucketAddress, error) {
	files, err := p.liveBucket.ListFiles("")
	if err != nil {
		return BackupBucketAddress{}, err
	}

	var executables []executor.Executable
	for _, file := range files {
		executables = append(executables, ExecutableBackup{file: file, backupAction: func(file string) error {
			return p.backupBucket.CopyObject(file, "", backupLocation, p.liveBucket.Name(), p.liveBucket.Region())
		}})
	}

	errs := p.executionStrategy.Run([][]executor.Executable{executables})
	if len(errs) != 0 {
		return BackupBucketAddress{}, formatErrors(
			fmt.Sprintf("failed to backup bucket %s", p.liveBucket.Name()),
			errs,
		)
	}

	return BackupBucketAddress{
		BucketName:   p.backupBucket.Name(),
		BucketRegion: p.backupBucket.Region(),
		Path:         backupLocation,
		EmptyBackup:  len(files) == 0,
	}, nil
}

type ExecutableBackup struct {
	file         string
	backupAction func(string) error
}

func (e ExecutableBackup) Execute() error {
	return e.backupAction(e.file)
}

func (p S3BucketPair) Restore(backupLocation string) error {
	files, err := p.backupBucket.ListFiles(backupLocation)
	if err != nil {
		return fmt.Errorf("cannot list files in %s\n %v", p.backupBucket.Name(), err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no files found in %s in bucket %s to restore", backupLocation, p.backupBucket.Name())
	}

	var executables []executor.Executable
	for _, file := range files {
		executables = append(executables, ExecutableBackup{file: file, backupAction: func(file string) error {
			return p.liveBucket.CopyObject(file, backupLocation, "", p.backupBucket.Name(), p.backupBucket.Region())
		}})
	}

	errs := p.executionStrategy.Run([][]executor.Executable{executables})
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
