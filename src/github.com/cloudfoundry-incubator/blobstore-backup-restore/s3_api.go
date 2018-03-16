package blobstore

import (
	"fmt"

	"math"

	"strings"

	"sort"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

const partSize int64 = 100 * 1024 * 1024

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

type partUploadOutput struct {
	completedPart *s3.CompletedPart
	err           error
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

func (s3Api S3Api) ListObjectVersions(bucketName string) ([]Version, error) {
	var versions []Version

	err := s3Api.s3Client.ListObjectVersionsPages(&s3.ListObjectVersionsInput{
		Bucket: aws.String(bucketName),
	}, func(output *s3.ListObjectVersionsOutput, lastPage bool) bool {
		for _, v := range output.Versions {
			fmt.Println(*v.Key)

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
		return nil, err
	}

	return versions, nil
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

	numParts := int64(math.Ceil(float64(blobSize) / float64(partSize)))
	partUploadOutputs := make(chan partUploadOutput, numParts)

	for i := int64(1); i <= numParts; i++ {
		go s3Api.partUpload(
			partUploadOutputs,
			i,
			blobSize,
			destinationBucketName,
			sourceBucketName,
			blobKey,
			versionId,
			*createOutput.UploadId,
		)
	}

	var parts []*s3.CompletedPart
	var errors []error
	for i := int64(0); i < numParts; i++ {
		partUploadOutput := <-partUploadOutputs
		if partUploadOutput.err != nil {
			errors = append(errors, partUploadOutput.err)
		} else {
			parts = append(parts, partUploadOutput.completedPart)
		}
	}

	if len(errors) != 0 {
		_, err := s3Api.s3Client.AbortMultipartUpload(&s3.AbortMultipartUploadInput{
			Bucket:   aws.String(destinationBucketName),
			Key:      aws.String(blobKey),
			UploadId: createOutput.UploadId,
		})
		if err != nil {
			errors = append(errors, err)
		}
		return formatErrors("errors occurred in multipart upload", errors)
	}

	sort.Slice(parts, func(i, j int) bool {
		return *parts[i].PartNumber < *parts[j].PartNumber
	})

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

func (s3Api S3Api) partUpload(
	partUploadOutputs chan<- partUploadOutput,
	partNumber,
	blobSize int64,
	destinationBucketName,
	sourceBucketName,
	blobKey,
	versionId,
	uploadId string) {

	partStart := (partNumber - 1) * partSize
	partEnd := int64(math.Min(float64(partNumber*partSize-1), float64(blobSize-1)))

	copyPartOutput, err := s3Api.s3Client.UploadPartCopy(&s3.UploadPartCopyInput{
		Bucket:          aws.String(destinationBucketName),
		Key:             aws.String(blobKey),
		UploadId:        aws.String(uploadId),
		CopySource:      aws.String(fmt.Sprintf("/%s/%s?versionId=%s", sourceBucketName, blobKey, versionId)),
		CopySourceRange: aws.String(fmt.Sprintf("bytes=%d-%d", partStart, partEnd)),
		PartNumber:      aws.Int64(partNumber),
	})

	if err != nil {
		partUploadOutputs <- partUploadOutput{
			&s3.CompletedPart{},
			fmt.Errorf("failed to upload part with range: %d-%d: %s", partStart, partEnd, err),
		}
	} else {
		partUploadOutputs <- partUploadOutput{
			&s3.CompletedPart{
				PartNumber: aws.Int64(partNumber),
				ETag:       copyPartOutput.CopyPartResult.ETag,
			},
			nil,
		}
	}
}

func (s3Api S3Api) DeleteObject(bucketName, file string) error {
	_, err := s3Api.s3Client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(file),
	})

	return err
}

func formatErrors(contextString string, errors []error) error {
	errorStrings := make([]string, len(errors))
	for i, err := range errors {
		errorStrings[i] = err.Error()
	}
	return fmt.Errorf("%s: %s", contextString, strings.Join(errorStrings, "\n"))
}
