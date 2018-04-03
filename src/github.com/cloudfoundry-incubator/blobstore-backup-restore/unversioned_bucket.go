package blobstore

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

//go:generate counterfeiter -o fakes/fake_unversioned_bucket.go . UnversionedBucket
type UnversionedBucket interface {
	Name() string
	RegionName() string
	Copy(key, destinationPath, originBucketName, originBucketRegion string) error
	ListFiles() ([]string, error)
}

type S3UnversionedBucket struct {
	S3Bucket
}

func NewS3UnversionedBucket(name, region, endpoint string, accessKey S3AccessKey) (S3UnversionedBucket, error) {
	s3Client, err := newS3Client(region, endpoint, accessKey)
	if err != nil {
		return S3UnversionedBucket{}, err
	}

	return S3UnversionedBucket{
		S3Bucket{
			name:       name,
			regionName: region,
			accessKey:  accessKey,
			endpoint:   endpoint,
			s3Client:   s3Client,
		},
	}, nil
}

func (bucket S3UnversionedBucket) ListFiles() ([]string, error) {
	var files []string
	err := bucket.s3Client.ListObjectsPages(&s3.ListObjectsInput{
		Bucket: aws.String(bucket.name),
	}, func(output *s3.ListObjectsOutput, lastPage bool) bool {
		for _, value := range output.Contents {
			files = append(files, *value.Key)
		}

		return !lastPage
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list files from bucket %s: %s", bucket.name, err)
	}
	return files, nil
}

func (bucket S3UnversionedBucket) Copy(blobKey, destinationPath, originBucketName, originBucketRegion string) error {

	blobSize, err := bucket.getBlobSize(originBucketName, originBucketRegion, blobKey, "null")
	if err != nil {
		return err
	}

	copySourceString := fmt.Sprintf("/%s/%s", originBucketName, blobKey)

	if sizeInMbs(blobSize) <= 100 {
		return bucket.copyVersionWithSingleRequest(originBucketName, blobKey, copySourceString, destinationPath)
	} else {
		return bucket.copyVersionWithMultipart(originBucketName, blobKey, copySourceString, destinationPath, blobSize)
	}
}
