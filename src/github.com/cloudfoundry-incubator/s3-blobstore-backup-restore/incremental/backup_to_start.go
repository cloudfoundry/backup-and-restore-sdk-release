package incremental

import "sort"

type BackupToStart struct {
	BucketPair            BackupBucketPair
	BackupDirectoryFinder BackupDirectoryFinder
	SameAsBucketID        string
}

func MarkSameBackupsToStart(backupsToStart map[string]BackupToStart) map[string]BackupToStart {
	liveBucketNamesToBucketIDs := make(map[string][]string)

	for bucketID, backupToStart := range backupsToStart {
		bucketIDs := liveBucketNamesToBucketIDs[backupToStart.BucketPair.ConfigLiveBucket.Name()]
		liveBucketNamesToBucketIDs[backupToStart.BucketPair.ConfigLiveBucket.Name()] = append(bucketIDs, bucketID)
	}

	markedBackupsToStart := make(map[string]BackupToStart)

	for _, bucketIDs := range liveBucketNamesToBucketIDs {
		sort.Strings(bucketIDs)

		for i, bucketID := range bucketIDs {
			if i == 0 {
				markedBackupsToStart[bucketID] = backupsToStart[bucketID]
			} else {
				markedBackupsToStart[bucketID] = BackupToStart{SameAsBucketID: bucketIDs[0]}
			}
		}
	}

	return markedBackupsToStart
}
