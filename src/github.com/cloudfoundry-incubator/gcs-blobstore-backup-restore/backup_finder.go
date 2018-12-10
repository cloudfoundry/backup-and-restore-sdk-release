package gcs

//go:generate counterfeiter -o fakes/fake_backup_finder.go . BackupFinder
type BackupFinder interface {
	ListBlobs() (map[string]Blob, error)
}

type LastBackupFinder struct {
	bucket Bucket
}

func NewLastBackupFinder(bucket Bucket) *LastBackupFinder {
	return &LastBackupFinder{
		bucket: bucket,
	}
}

func (f *LastBackupFinder) ListBlobs() (map[string]Blob, error) {
	return nil, nil
}
