package blobstore

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

const partSize int64 = 100 * 1024 * 1024

//go:generate counterfeiter -o fakes/fake_versioned_bucket.go . VersionedBucket
type VersionedBucket interface {
	Name() string
	RegionName() string
	Versions() ([]Version, error)
	CopyVersions(regionName, bucketName string, versions []BlobVersion) error
}

type Version struct {
	Key      string
	Id       string `json:"VersionId"`
	IsLatest bool
}

type S3VersionedBucket struct {
	S3Bucket
}

func NewS3VersionedBucket(name, region, endpoint string, accessKey S3AccessKey) (S3VersionedBucket, error) {
	s3Client, err := newS3Client(region, endpoint, accessKey)
	if err != nil {
		return S3VersionedBucket{}, err
	}

	return S3VersionedBucket{
		S3Bucket{
			name:       name,
			regionName: region,
			accessKey:  accessKey,
			endpoint:   endpoint,
			s3Client:   s3Client,
		},
	}, nil
}

func (bucket S3VersionedBucket) Versions() ([]Version, error) {
	err := bucket.checkIfVersioned()
	if err != nil {
		return nil, err
	}

	var versions []Version
	err = bucket.s3Client.ListObjectVersionsPages(&s3.ListObjectVersionsInput{
		Bucket: aws.String(bucket.name),
	}, func(output *s3.ListObjectVersionsOutput, lastPage bool) bool {
		for _, v := range output.Versions {
			version := Version{
				Key:      *v.Key,
				Id:       *v.VersionId,
				IsLatest: *v.IsLatest,
			}
			versions = append(versions, version)
		}

		return !lastPage
	})

	if err != nil {
		return nil, fmt.Errorf("failed to retrieve versions from bucket %s: %s", bucket.name, err)
	}

	return versions, nil
}

func (bucket S3VersionedBucket) CopyVersions(sourceBucketRegion, sourceBucketName string, versionsToCopy []BlobVersion) error {
	err := bucket.checkIfVersioned()
	if err != nil {
		return err
	}

	for _, versionToCopy := range versionsToCopy {
		err = bucket.copyVersion(sourceBucketName, sourceBucketRegion, versionToCopy.BlobKey, versionToCopy.Id, "")
		if err != nil {

			return err
		}
	}

	return nil
}

func (bucket S3VersionedBucket) checkIfVersioned() error {
	output, err := bucket.s3Client.GetBucketVersioning(&s3.GetBucketVersioningInput{
		Bucket: &bucket.name,
	})

	if err != nil {
		return fmt.Errorf("could not check if bucket %s is versioned: %s", bucket.name, err)
	}

	if output == nil || output.Status == nil || *output.Status != "Enabled" {
		return fmt.Errorf("bucket %s is not versioned", bucket.name)
	}

	return nil
}

func (bucket S3VersionedBucket) copyVersion(sourceBucketName, sourceBucketRegion, blobKey, versionId, destinationPath string) error {
	blobSize, err := bucket.getBlobSize(sourceBucketName, sourceBucketRegion, blobKey, versionId)
	if err != nil {
		return err
	}
	copySourceString := fmt.Sprintf("/%s/%s?versionId=%s", sourceBucketName, blobKey, versionId)

	if sizeInMbs(blobSize) <= 100 {
		return bucket.copyVersionWithSingleRequest(sourceBucketName, blobKey, copySourceString, destinationPath)
	} else {
		return bucket.copyVersionWithMultipart(sourceBucketName, blobKey, copySourceString, destinationPath, blobSize)
	}
}
