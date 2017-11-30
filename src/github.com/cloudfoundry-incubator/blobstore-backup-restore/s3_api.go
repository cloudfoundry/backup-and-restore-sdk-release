package blobstore

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func NewS3API(endpoint, regionName, accessKeyId, accessKeySecret string) (S3Api, error) {
	awsSession, err := session.NewSession(&aws.Config{
		Region:      &regionName,
		Credentials: credentials.NewStaticCredentials(accessKeyId, accessKeySecret, ""),
		Endpoint:    aws.String(endpoint),
	})

	if err != nil {
		return S3Api{}, err
	}

	s3Session := s3.New(awsSession)
	return S3Api{regionName: regionName, S3: s3Session}, nil
}

type S3Api struct {
	regionName string
	*s3.S3
}

func (s3Api S3Api) CopyVersion(sourceBucketName, blobKey, versionId, destinationBucketName string) error {
	input := s3.CopyObjectInput{
		Bucket:     aws.String(destinationBucketName),
		Key:        aws.String(blobKey),
		CopySource: aws.String(fmt.Sprintf("/%s/%s?versionId=%s", sourceBucketName, blobKey, versionId)),
	}

	_, err := s3Api.S3.CopyObject(&input)

	return err
}

func (s3Api S3Api) DeleteObject(bucketName, file string) error {
	input := s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(file),
	}

	_, err := s3Api.S3.DeleteObject(&input)

	return err
}
