package incremental

import "fmt"

//go:generate counterfeiter -o fakes/fake_backup_finder.go . BackupFinder
type BackupFinder interface {
	Find() (BackupDirectory, error)
}

type BucketPair struct {
	LiveBucket   Bucket
	BackupBucket Bucket
}

type BackupStarter struct {
	BucketPair   BucketPair
	BackupFinder BackupFinder
}

func (b BackupStarter) Run() error {
	// find the last complete backup and list blobs
	_, err := b.BackupFinder.Find()
	if err != nil {
		return fmt.Errorf("failed to start backup: %s", err)
	}

	// list blobs in the live bucket

	// create a new backup directory

	// copy new live blobs to the new backup directory

	// write the backup artifact for restore

	// write the backup directory and list of previously backed up blobs for completer

	return nil
}
