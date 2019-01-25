package incremental

import "regexp"

type BackupDirectoryFinder struct {
	ID     string
	Bucket Bucket
}

func (b BackupDirectoryFinder) ListBlobs(bucketID string) ([]BackedUpBlob, error) {
	// list directories
	dirs, _ := b.Bucket.ListDirectories()

	// filter for backup directories
	regex := regexp.MustCompile(`^\d{4}(_\d{2}){5}$`)

	var filteredDirs []string
	for _, dir := range dirs {
		if regex.MatchString(dir) {
			filteredDirs = append(filteredDirs, dir)
		}
	}

	// if none, return empty list of blobs
	if len(filteredDirs) == 0 {
		return nil, nil
	}

	// identify last complete backup directory

	// if none, return empty list of blobs

	// list blobs in last complete backup directory
	backupDirPath := joinBlobPath(filteredDirs[0], b.ID)
	blobs, _ := b.Bucket.ListBlobs(backupDirPath)

	var backedUpBlobs []BackedUpBlob
	for _, blob := range blobs {
		backedUpBlobs = append(backedUpBlobs, BackedUpBlob{
			Path:                blob.Path(),
			BackupDirectoryPath: backupDirPath,
		})
	}

	return backedUpBlobs, nil
}
