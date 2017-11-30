package blobstore

import (
	"encoding/json"
	"strings"
)

//go:generate counterfeiter -o fakes/fake_bucket.go . Bucket
type Bucket interface {
	Name() string
	RegionName() string
	Versions() ([]Version, error)
	CopyVersionsAndPrune(regionName, bucketName string, versions []BlobVersion) error
}

type S3Bucket struct {
	awsCliPath string
	name       string
	regionName string
	accessKey  S3AccessKey
	endpoint   string
}

type S3AccessKey struct {
	Id     string
	Secret string
}

func NewS3Bucket(awsCliPath, name, region, endpoint string, accessKey S3AccessKey) S3Bucket {
	return S3Bucket{
		awsCliPath: awsCliPath,
		name:       name,
		regionName: region,
		accessKey:  accessKey,
		endpoint:   endpoint,
	}
}

func (bucket S3Bucket) Name() string {
	return bucket.name
}

func (bucket S3Bucket) RegionName() string {
	return bucket.regionName
}

func (bucket S3Bucket) Versions() ([]Version, error) {
	s3Cli := NewS3CLI(bucket.awsCliPath, bucket.endpoint, bucket.regionName, bucket.accessKey.Id, bucket.accessKey.Secret)

	output, err := s3Cli.ListObjectVersions(bucket.name)

	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(string(output)) == "" {
		return []Version{}, nil
	}

	response := S3ListVersionsResponse{}
	err = json.Unmarshal(output, &response)
	if err != nil {
		return nil, err
	}

	return response.Versions, nil
}

func (bucket S3Bucket) CopyVersionsAndPrune(sourceBucketRegion, sourceBucketName string, versionsToCopy []BlobVersion) error {
	var err error

	s3Cli := NewS3CLI(bucket.awsCliPath, bucket.endpoint, bucket.regionName, bucket.accessKey.Id, bucket.accessKey.Secret)
	s3API, err := NewS3API(bucket.endpoint, bucket.regionName, bucket.accessKey.Id, bucket.accessKey.Secret)

	if err != nil {
		return err
	}

	for _, versionToCopy := range versionsToCopy {
		err = s3API.CopyVersion(sourceBucketName, versionToCopy.BlobKey, versionToCopy.Id, bucket.name)
		if err != nil {
			return err
		}
	}

	var keysThatShouldBePresentInBucket []string
	for _, versionToCopy := range versionsToCopy {
		keysThatShouldBePresentInBucket = append(keysThatShouldBePresentInBucket, versionToCopy.BlobKey)
	}

	keysThatArePresentInBucket, err := s3Cli.ListObjects(bucket.name)
	if err != nil {
		return err
	}

	for _, key := range keysThatArePresentInBucket {
		if !includes(keysThatShouldBePresentInBucket, key) {
			err = s3API.DeleteObject(bucket.name, key)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
func includes(haystack []string, needle string) bool {
	for _, value := range haystack {
		if value == needle {
			return true
		}
	}
	return false
}

type S3ListVersionsResponse struct {
	Versions []Version
}

type Version struct {
	Key      string
	Id       string `json:"VersionId"`
	IsLatest bool
}

type S3ListResponse struct {
	Contents []Object
}

type Object struct {
	Key string
}
