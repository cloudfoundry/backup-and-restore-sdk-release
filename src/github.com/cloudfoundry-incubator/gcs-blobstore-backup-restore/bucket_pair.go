package gcs

import (
	"errors"
	"fmt"

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

	filteredConfig, err := filterConfig(config)
	if err != nil {
		return nil, err
	}

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

func filterConfig(originalConfig map[string]Config) (map[string]Config, error) {
	filteredConfig := make(map[string]Config)

	for name, config := range originalConfig {
		key := findKeyWithConfig(filteredConfig, config)
		errorMessage := "cannot use reserved bucket pair name: %s"
		if key == "" {
			_, exists := filteredConfig[name]
			if exists {
				return nil, errors.New(fmt.Sprintf(errorMessage, name))
			}
			filteredConfig[name] = config

		} else {
			mergedKeyName := name + "-" + key
			_, exists := filteredConfig[mergedKeyName]
			if exists {
				return nil, errors.New(fmt.Sprintf(errorMessage, mergedKeyName))
			}
			filteredConfig[mergedKeyName] = config

			delete(filteredConfig, key)
		}
	}
	return filteredConfig, nil
}

func findKeyWithConfig(configs map[string]Config, expectedConfig Config) string {
	for name, config := range configs {
		if config == expectedConfig {
			return name
		}
	}
	return ""
}
