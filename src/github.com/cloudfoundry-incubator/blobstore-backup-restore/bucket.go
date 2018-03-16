package blobstore

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

const partSize int64 = 100 * 1024 * 1024

//go:generate counterfeiter -o fakes/fake_bucket.go . Bucket
type Bucket interface {
	Name() string
	RegionName() string
	Versions() ([]Version, error)
	CopyVersions(regionName, bucketName string, versions []BlobVersion) error
}

type S3Bucket struct {
	name       string
	regionName string
	accessKey  S3AccessKey
	endpoint   string
	s3Client   *s3.S3
}

type S3AccessKey struct {
	Id     string
	Secret string
}

type Version struct {
	Key      string
	Id       string `json:"VersionId"`
	IsLatest bool
}

func NewS3Bucket(name, region, endpoint string, accessKey S3AccessKey) (S3Bucket, error) {
	s3Client, err := newS3Client(region, endpoint, accessKey)
	if err != nil {
		return S3Bucket{}, err
	}

	return S3Bucket{
		name:       name,
		regionName: region,
		accessKey:  accessKey,
		endpoint:   endpoint,
		s3Client:   s3Client,
	}, nil
}

func (bucket S3Bucket) Name() string {
	return bucket.name
}

func (bucket S3Bucket) RegionName() string {
	return bucket.regionName
}

func (bucket S3Bucket) Versions() ([]Version, error) {
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

func (bucket S3Bucket) CopyVersions(sourceBucketRegion, sourceBucketName string, versionsToCopy []BlobVersion) error {
	err := bucket.checkIfVersioned()
	if err != nil {
		return err
	}

	for _, versionToCopy := range versionsToCopy {
		err = bucket.copyVersion(sourceBucketName, sourceBucketRegion, versionToCopy.BlobKey, versionToCopy.Id, bucket.name)
		if err != nil {
			return err
		}
	}

	return nil
}

func (bucket S3Bucket) checkIfVersioned() error {
	output, err := bucket.s3Client.GetBucketVersioning(&s3.GetBucketVersioningInput{
		Bucket: &bucket.name,
	})

	if err != nil {
		return fmt.Errorf("failed to check if bucket %s is versioned: %s", bucket.name, err)
	}

	if output == nil || output.Status == nil || *output.Status != "Enabled" {
		return fmt.Errorf("bucket %s is not versioned", bucket.name)
	}

	return nil
}

func (bucket S3Bucket) copyVersion(sourceBucketName, sourceBucketRegion, blobKey, versionId, destinationBucketName string) error {
	blobSize, err := bucket.getBlobSize(sourceBucketName, sourceBucketRegion, blobKey, versionId)
	if err != nil {
		return err
	}

	if sizeInMbs(blobSize) <= 100 {
		return bucket.copyVersionWithSingleRequest(sourceBucketName, blobKey, versionId, destinationBucketName)
	} else {
		return bucket.copyVersionWithMultipart(sourceBucketName, blobKey, versionId, destinationBucketName, blobSize)
	}
}

func (bucket S3Bucket) getBlobSize(bucketName, bucketRegion, blobKey, versionId string) (int64, error) {
	s3Client, err := newS3Client(bucketRegion, bucket.endpoint, bucket.accessKey)
	if err != nil {
		return 0, err
	}

	headObjectOutput, err := s3Client.HeadObject(&s3.HeadObjectInput{
		Bucket:    aws.String(bucketName),
		Key:       aws.String(blobKey),
		VersionId: aws.String(versionId),
	})

	if err != nil {
		return 0, fmt.Errorf("failed to retrieve head object output: %s", err)
	}

	return *headObjectOutput.ContentLength, nil
}

func sizeInMbs(sizeInBytes int64) int64 {
	return sizeInBytes / (1024 * 1024)
}

func (bucket S3Bucket) copyVersionWithSingleRequest(sourceBucketName, blobKey, versionId, destinationBucketName string) error {
	_, err := bucket.s3Client.CopyObject(&s3.CopyObjectInput{
		Bucket:     aws.String(destinationBucketName),
		Key:        aws.String(blobKey),
		CopySource: aws.String(fmt.Sprintf("/%s/%s?versionId=%s", sourceBucketName, blobKey, versionId)),
	})

	return err
}

func (bucket S3Bucket) copyVersionWithMultipart(sourceBucketName, blobKey, versionId, destinationBucketName string, blobSize int64) error {
	createOutput, err := bucket.s3Client.CreateMultipartUpload(&s3.CreateMultipartUploadInput{
		Bucket: aws.String(destinationBucketName),
		Key:    aws.String(blobKey),
	})

	if err != nil {
		return fmt.Errorf("failed to create multipart upload: %s", err)
	}

	numParts := int64(math.Ceil(float64(blobSize) / float64(partSize)))
	partUploadOutputs := make(chan partUploadOutput, numParts)

	for i := int64(1); i <= numParts; i++ {
		go func(partNumber int64) {
			partStart := (partNumber - 1) * partSize
			partEnd := int64(math.Min(float64(partNumber*partSize-1), float64(blobSize-1)))

			copyPartOutput, err := bucket.s3Client.UploadPartCopy(&s3.UploadPartCopyInput{
				Bucket:          aws.String(destinationBucketName),
				Key:             aws.String(blobKey),
				UploadId:        aws.String(*createOutput.UploadId),
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
		}(i)
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
		_, err := bucket.s3Client.AbortMultipartUpload(&s3.AbortMultipartUploadInput{
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

	_, err = bucket.s3Client.CompleteMultipartUpload(&s3.CompleteMultipartUploadInput{
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

type partUploadOutput struct {
	completedPart *s3.CompletedPart
	err           error
}

func formatErrors(contextString string, errors []error) error {
	errorStrings := make([]string, len(errors))
	for i, err := range errors {
		errorStrings[i] = err.Error()
	}
	return fmt.Errorf("%s: %s", contextString, strings.Join(errorStrings, "\n"))
}

func newS3Client(regionName string, endpoint string, accessKey S3AccessKey) (*s3.S3, error) {
	awsSession, err := session.NewSession(&aws.Config{
		Region:           &regionName,
		Credentials:      credentials.NewStaticCredentials(accessKey.Id, accessKey.Secret, ""),
		Endpoint:         aws.String(endpoint),
		S3ForcePathStyle: aws.Bool(true),
	})

	if err != nil {
		return nil, err
	}

	return s3.New(awsSession), nil
}
