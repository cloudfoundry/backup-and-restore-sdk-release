package config

import "github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/s3"

type BucketConfig struct {
	Name               string `json:"name"`
	Region             string `json:"region"`
	AwsAccessKeyId     string `json:"aws_access_key_id"`
	AwsSecretAccessKey string `json:"aws_secret_access_key"`
	Endpoint           string `json:"endpoint"`
	UseIAMProfile      bool   `json:"use_iam_profile"`
}

func BuildVersionedBuckets(config map[string]BucketConfig) (map[string]s3.VersionedBucket, error) {
	var buckets = map[string]s3.VersionedBucket{}

	for identifier, bucketConfig := range config {
		s3Bucket, err := s3.NewBucket(
			bucketConfig.Name,
			bucketConfig.Region,
			bucketConfig.Endpoint,
			s3.AccessKey{
				Id:     bucketConfig.AwsAccessKeyId,
				Secret: bucketConfig.AwsSecretAccessKey,
			},
			bucketConfig.UseIAMProfile,
		)
		if err != nil {
			return nil, err
		}

		buckets[identifier] = s3Bucket
	}

	return buckets, nil
}
