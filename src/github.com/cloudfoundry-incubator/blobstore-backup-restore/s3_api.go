package blobstore

import (
	"fmt"

	"math"

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

func (s3Api S3Api) GetBlobSize(bucketName, blobKey, versionId string) (int64, error) {
	headObjectOutput, err := s3Api.s3Client.HeadObject(&s3.HeadObjectInput{
		Bucket:    aws.String(bucketName),
		Key:       aws.String(blobKey),
		VersionId: aws.String(versionId),
	})

	if err != nil {
		return 0, fmt.Errorf("failed to retrieve head object output: %s", err)
	}

	return *headObjectOutput.ContentLength, nil
}

func (s3Api S3Api) CopyVersion(sourceBucketName, blobKey, versionId, destinationBucketName string, blobSize int64) error {
	if sizeInMbs(blobSize) <= 100 {
		return s3Api.copyVersion(sourceBucketName, blobKey, versionId, destinationBucketName)
	} else {
		return s3Api.copyVersionWithMultipart(sourceBucketName, blobKey, versionId, destinationBucketName, blobSize)
	}
}

func sizeInMbs(sizeInBytes int64) int64 {
	return sizeInBytes / (1024 * 1024)
}

func (s3Api S3Api) copyVersion(sourceBucketName, blobKey, versionId, destinationBucketName string) error {
	_, err := s3Api.s3Client.CopyObject(&s3.CopyObjectInput{
		Bucket:     aws.String(destinationBucketName),
		Key:        aws.String(blobKey),
		CopySource: aws.String(fmt.Sprintf("/%s/%s?versionId=%s", sourceBucketName, blobKey, versionId)),
	})

	return err
}

func (s3Api S3Api) copyVersionWithMultipart(sourceBucketName, blobKey, versionId, destinationBucketName string, blobSize int64) error {
	createOutput, err := s3Api.s3Client.CreateMultipartUpload(&s3.CreateMultipartUploadInput{
		Bucket: aws.String(destinationBucketName),
		Key:    aws.String(blobKey),
	})

	if err != nil {
		return fmt.Errorf("failed to create multipart upload: %s", err)
	}

	var partSize int64 = 10 * 1024 * 1024
	var partNumber int64 = 1
	var parts []*s3.CompletedPart
	for i := int64(0); i < blobSize; i += partSize {
		upperLimit := int64(math.Min(float64(i+partSize-1), float64(blobSize)-1))

		copyPartOutput, err := s3Api.s3Client.UploadPartCopy(&s3.UploadPartCopyInput{
			Bucket:          aws.String(destinationBucketName),
			Key:             aws.String(blobKey),
			UploadId:        createOutput.UploadId,
			CopySource:      aws.String(fmt.Sprintf("/%s/%s?versionId=%s", sourceBucketName, blobKey, versionId)),
			CopySourceRange: aws.String(fmt.Sprintf("bytes=%d-%d", i, upperLimit)),
			PartNumber:      aws.Int64(partNumber),
		})

		if err != nil {
			return fmt.Errorf("failed to upload part with range: %d-%d: %s", i, upperLimit, err)
		}

		parts = append(parts, &s3.CompletedPart{
			PartNumber: aws.Int64(partNumber),
			ETag:       copyPartOutput.CopyPartResult.ETag,
		})

		partNumber++
	}

	_, err = s3Api.s3Client.CompleteMultipartUpload(&s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(destinationBucketName),
		Key:      aws.String(blobKey),
		UploadId: createOutput.UploadId,
		MultipartUpload: &s3.CompletedMultipartUpload{
			Parts: parts,
		},
	})

	if err != nil {
		return fmt.Errorf("failed to complete multipart upload: %s", err)
	}

	return nil
}

func (s3Api S3Api) DeleteObject(bucketName, file string) error {
	_, err := s3Api.s3Client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(file),
	})

	return err
}
