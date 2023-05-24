package versioned

import (
	"s3-blobstore-backup-restore/s3bucket"
)

type BucketConfig struct {
	Name               string `json:"name"`
	Region             string `json:"region"`
	AwsAccessKeyId     string `json:"aws_access_key_id"`
	AwsSecretAccessKey string `json:"aws_secret_access_key"`
	// # Warning
	//
	// AwsAssumedRoleArn is provided as is and isn't thoroughly tested
	AwsAssumedRoleArn string `json:"aws_assumed_role_arn,omitempty"`
	Endpoint          string `json:"endpoint"`
	UseIAMProfile     bool   `json:"use_iam_profile"`
	ForcePathStyle    bool   `json:"force_path_style"`
}

type NewBucket func(bucketName, bucketRegion, endpoint, role string, accessKey s3bucket.AccessKey, useIAMProfile, forcePathStyle bool) (Bucket, error)

func BuildVersionedBuckets(config map[string]BucketConfig, newbucket NewBucket) (map[string]Bucket, error) {
	var buckets = map[string]Bucket{}

	for identifier, bucketConfig := range config {
		s3Bucket, err := newbucket(
			bucketConfig.Name,
			bucketConfig.Region,
			bucketConfig.Endpoint,
			bucketConfig.AwsAssumedRoleArn,
			s3bucket.AccessKey{
				Id:     bucketConfig.AwsAccessKeyId,
				Secret: bucketConfig.AwsSecretAccessKey,
			},
			bucketConfig.UseIAMProfile,
			bucketConfig.ForcePathStyle,
		)
		if err != nil {
			return nil, err
		}

		buckets[identifier] = s3Bucket
	}

	return buckets, nil
}
