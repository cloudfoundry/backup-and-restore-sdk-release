package incremental

import (
	"regexp"
	"sort"
)

//go:generate counterfeiter -o fakes/fake_backup_directory_finder.go . BackupDirectoryFinder
type BackupDirectoryFinder interface {
	ListBlobs(string, Bucket) ([]BackedUpBlob, error)
}

type Finder struct{}

func (b Finder) ListBlobs(bucketID string, backupBucket Bucket) ([]BackedUpBlob, error) {
	dirs, err := backupBucket.ListDirectories()
	if err != nil {
		return nil, err
	}

	regex := regexp.MustCompile(`^\d{4}(_\d{2}){5}$`)

	var filteredDirs []string
	for _, dir := range dirs {
		if regex.MatchString(dir) {
			filteredDirs = append(filteredDirs, dir)
		}
	}

	if len(filteredDirs) == 0 {
		return nil, nil
	}

	sort.Sort(sort.Reverse(sort.StringSlice(filteredDirs)))

	var backupDirs []BackupDirectory
	for _, filteredDir := range filteredDirs {
		backupDirs = append(backupDirs, BackupDirectory{
			Path:   joinBlobPath(filteredDir, bucketID),
			Bucket: backupBucket,
		})
	}

	lastCompleteBackupDir, err := b.findLastCompleteBackup(backupDirs)
	if err != nil {
		return nil, err
	}

	if lastCompleteBackupDir == nil {
		return nil, nil
	}

	blobs, err := backupBucket.ListBlobs(lastCompleteBackupDir.Path)
	if err != nil {
		return nil, err
	}

	var backedUpBlobs []BackedUpBlob
	for _, blob := range blobs {
		backedUpBlobs = append(backedUpBlobs, BackedUpBlob{
			Path:                blob.Path(),
			BackupDirectoryPath: lastCompleteBackupDir.Path,
		})
	}

	return backedUpBlobs, nil
}

func (b Finder) findLastCompleteBackup(dirs []BackupDirectory) (*BackupDirectory, error) {
	for _, dir := range dirs {
		isComplete, err := dir.IsComplete()
		if err != nil {
			return nil, err
		}

		if isComplete {
			return &dir, nil
		}
	}

	return nil, nil
}
