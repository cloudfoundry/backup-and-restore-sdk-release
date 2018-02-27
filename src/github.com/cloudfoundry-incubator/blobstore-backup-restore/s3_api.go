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

	s3Client := s3.New(awsSession)
	return S3Api{regionName: regionName, s3Client: s3Client}, nil
}

type S3Api struct {
	regionName string
	s3Client   *s3.S3
}

func (s3Api S3Api) IsVersioned(bucketName string) (bool, error) {
	output, err := s3Api.s3Client.GetBucketVersioning(&s3.GetBucketVersioningInput{
		Bucket: &bucketName,
	})

	if err != nil {
		return false, err
	}

	return output != nil && output.Status != nil && *output.Status == "Enabled", nil
}

func (s3Api S3Api) CopyVersion(sourceBucketName, blobKey, versionId, destinationBucketName string) error {
	_, err := s3Api.s3Client.CopyObject(&s3.CopyObjectInput{
		Bucket:     aws.String(destinationBucketName),
		Key:        aws.String(blobKey),
		CopySource: aws.String(fmt.Sprintf("/%s/%s?versionId=%s", sourceBucketName, blobKey, versionId)),
	})

	return err
}

func (s3Api S3Api) DeleteObject(bucketName, file string) error {
	_, err := s3Api.s3Client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(file),
	})

	return err
}
