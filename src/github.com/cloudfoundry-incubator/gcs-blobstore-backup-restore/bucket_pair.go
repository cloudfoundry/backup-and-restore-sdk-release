package gcs

import (
	"sort"

	"github.com/cloudfoundry-incubator/bosh-backup-and-restore/executor"
)

type BucketPair struct {
	LiveBucket        Bucket
	BackupBucket      Bucket
	ExecutionStrategy executor.Executor
}

func BuildBackupsToComplete(gcpServiceAccountKey string, config map[string]Config) (map[string]BackupToComplete, error) {
	backupsToComplete := map[string]BackupToComplete{}
	exe := executor.NewParallelExecutor()
	exe.SetMaxInFlight(200)
	for bucketID, bucketConfig := range config {
		bucket, err := NewSDKBucket(gcpServiceAccountKey, bucketConfig.BucketName)
		if err != nil {
			return nil, err
		}

		backupBucket, err := NewSDKBucket(gcpServiceAccountKey, bucketConfig.BackupBucketName)
		if err != nil {
			return nil, err
		}

		bucketPair := BucketPair{
			LiveBucket:        bucket,
			BackupBucket:      backupBucket,
			ExecutionStrategy: exe,
		}
		backupsToComplete[bucketID] = BackupToComplete{
			BucketPair:     bucketPair,
			SameAsBucketID: "",
		}
	}

	markedSameBackupsToComplete := MarkSameBackupsToComplete(backupsToComplete)

	return markedSameBackupsToComplete, nil
}

func MarkSameBackupsToComplete(backupsToComplete map[string]BackupToComplete) map[string]BackupToComplete {
	liveBucketNamesToBucketIDs := make(map[string][]string)

	for bucketID, backupToComplete := range backupsToComplete {
		bucketIDs := liveBucketNamesToBucketIDs[backupToComplete.BucketPair.LiveBucket.Name()]
		liveBucketNamesToBucketIDs[backupToComplete.BucketPair.LiveBucket.Name()] = append(bucketIDs, bucketID)
	}

	markedSameBackupsToComplete := make(map[string]BackupToComplete)

	for _, bucketIDs := range liveBucketNamesToBucketIDs {
		sort.Strings(bucketIDs)

		for i, bucketID := range bucketIDs {
			if i == 0 {
				markedSameBackupsToComplete[bucketID] = backupsToComplete[bucketID]
			} else {
				markedSameBackupsToComplete[bucketID] = BackupToComplete{SameAsBucketID: bucketIDs[0]}
			}
		}
	}

	return markedSameBackupsToComplete
}
