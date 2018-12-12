package gcs

type BucketPair struct {
	LiveBucket   Bucket
	BackupBucket Bucket
	BackupFinder BackupFinder
}

func BuildBucketPairs(gcpServiceAccountKey string, config map[string]Config) (map[string]BucketPair, error) {
	buckets := map[string]BucketPair{}

	for bucketID, bucketConfig := range config {
		bucket, err := NewSDKBucket(gcpServiceAccountKey, bucketConfig.BucketName)
		if err != nil {
			return nil, err
		}

		backupBucket, err := NewSDKBucket(gcpServiceAccountKey, bucketConfig.BackupBucketName)
		if err != nil {
			return nil, err
		}

		buckets[bucketID] = BucketPair{
			LiveBucket:   bucket,
			BackupBucket: backupBucket,
			BackupFinder: NewLastBackupFinder(backupBucket),
		}
	}

	return buckets, nil
}
