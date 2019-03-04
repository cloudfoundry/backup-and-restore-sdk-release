package gcs

import "github.com/cloudfoundry-incubator/bosh-backup-and-restore/executor"

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

	return backupsToComplete, nil
}
