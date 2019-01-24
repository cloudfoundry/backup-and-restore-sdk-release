package gcs

import (
	"github.com/cloudfoundry-incubator/bosh-backup-and-restore/executor"
)

type BucketPair struct {
	LiveBucket        Bucket
	BackupBucket      Bucket
	ExecutionStrategy executor.Executor
}

func BuildBucketPairs(gcpServiceAccountKey string, config map[string]Config) (map[string]BucketPair, error) {
	buckets := map[string]BucketPair{}
	exe := executor.NewParallelExecutor()
	exe.SetMaxInFlight(200)

	filteredConfig := filterConfig(config)

	for bucketPairName, bucketConfig := range filteredConfig {
		bucket, err := NewSDKBucket(gcpServiceAccountKey, bucketConfig.BucketName)
		if err != nil {
			return nil, err
		}

		backupBucket, err := NewSDKBucket(gcpServiceAccountKey, bucketConfig.BackupBucketName)
		if err != nil {
			return nil, err
		}

		buckets[bucketPairName] = BucketPair{
			LiveBucket:        bucket,
			BackupBucket:      backupBucket,
			ExecutionStrategy: exe,
		}
	}

	return buckets, nil
}

func filterConfig(originalConfig map[string]Config) map[string]Config {
	filteredConfig := make(map[string]Config)

	for name, config := range originalConfig {
		key := findKeyWithConfig(filteredConfig, config)
		if key == "" {
			filteredConfig[name] = config
		} else {
			filteredConfig[name+"-"+key] = config
			delete(filteredConfig, key)
		}
	}
	return filteredConfig
}

func findKeyWithConfig(configs map[string]Config, expectedConfig Config) string {
	for name, config := range configs {
		if config == expectedConfig {
			return name
		}
	}
	return ""
}
