package s3

import (
	"fmt"
	"math"
	"sort"

	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

const partSize int64 = 100 * 1024 * 1024

type Bucket struct {
	name          string
	regionName    string
	accessKey     AccessKey
	endpoint      string
	s3Client      *s3.S3
	useIAMProfile bool
}

type AccessKey struct {
	Id     string
	Secret string
}

type Version struct {
	Key      string
	Id       string `json:"VersionId"`
	IsLatest bool
}

//go:generate counterfeiter -o fakes/fake_unversioned_bucket.go . UnversionedBucket
type UnversionedBucket interface {
	Name() string
	RegionName() string
	CopyObject(key, originPath, destinationPath, originBucketName, originBucketRegion string) error
	ListFiles(path string) ([]string, error)
}

//go:generate counterfeiter -o fakes/fake_versioned_bucket.go . VersionedBucket
type VersionedBucket interface {
	Name() string
	RegionName() string
	CopyVersion(blobKey, versionId, originBucketName, originBucketRegion string) error
	ListVersions() ([]Version, error)
	CheckIfVersioned() error
}

func NewBucket(bucketName, bucketRegion, endpoint string, accessKey AccessKey, useIAMProfile bool) (Bucket, error) {
	s3Client, err := newS3Client(bucketRegion, endpoint, accessKey, useIAMProfile)
	if err != nil {
		return Bucket{}, err
	}

	return Bucket{
		name:          bucketName,
		regionName:    bucketRegion,
		s3Client:      s3Client,
		accessKey:     accessKey,
		endpoint:      endpoint,
		useIAMProfile: useIAMProfile,
	}, nil
}

func (bucket Bucket) Name() string {
	return bucket.name
}

func (bucket Bucket) RegionName() string {
	return bucket.regionName
}

func (bucket Bucket) ListFiles(subfolder string) ([]string, error) {
	var files []string
	err := bucket.s3Client.ListObjectsPages(&s3.ListObjectsInput{
		Bucket: aws.String(bucket.name),
		Prefix: aws.String(subfolder),
	}, func(output *s3.ListObjectsOutput, lastPage bool) bool {
		for _, value := range output.Contents {
			files = append(files, *value.Key)
		}

		return !lastPage
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list files from bucket %s: %s", bucket.name, err)
	}

	if subfolder != "" {
		for index, fileLocation := range files {
			files[index] = strings.Replace(fileLocation, subfolder+"/", "", 1)
		}
	}

	return files, nil
}

func (bucket Bucket) ListVersions() ([]Version, error) {
	err := bucket.CheckIfVersioned()
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

func (bucket Bucket) CheckIfVersioned() error {
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

func (bucket Bucket) CopyObject(blobKey, originPath, destinationPath, originBucketName, originBucketRegion string) error {
	return bucket.copyVersion(
		blobKey,
		"null",
		originPath,
		destinationPath,
		originBucketName,
		originBucketRegion,
	)
}

func (bucket Bucket) CopyVersion(blobKey, versionId, originBucketName, originBucketRegion string) error {
	return bucket.copyVersion(
		blobKey,
		versionId,
		"",
		"",
		originBucketName,
		originBucketRegion,
	)
}

func (bucket Bucket) copyVersion(blobKey, versionId, originPath, destinationPath, originBucketName, originBucketRegion string) error {
	blobSize, err := bucket.getBlobSize(originBucketName, originBucketRegion, originPath, blobKey, versionId)
	if err != nil {
		return err
	}

	copySourceString := fmt.Sprintf("/%s/%s/%s?versionId=%s", originBucketName, originPath, blobKey, versionId)
	copySourceString = strings.Replace(copySourceString, "//", "/", -1)

	if sizeInMbs(blobSize) <= 1024 {
		return bucket.copyVersionWithSingleRequest(originBucketName, blobKey, copySourceString, destinationPath)
	} else {
		return bucket.copyVersionWithMultipart(originBucketName, blobKey, copySourceString, destinationPath, blobSize)
	}
}

func (bucket Bucket) getBlobSize(bucketName, bucketRegion, originPath, blobKey, versionId string) (int64, error) {
	s3Client, err := newS3Client(bucketRegion, bucket.endpoint, bucket.accessKey, bucket.useIAMProfile)
	if err != nil {
		return 0, err
	}

	headObjectOutput, err := s3Client.HeadObject(&s3.HeadObjectInput{
		Bucket:    aws.String(bucketName),
		Key:       aws.String(fmt.Sprintf("%s/%s", originPath, blobKey)),
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

func (bucket Bucket) copyVersionWithSingleRequest(sourceBucketName, blobKey, copySourceString, destinationPath string) error {
	destinationKey := fmt.Sprintf("%s/%s", destinationPath, blobKey)
	_, err := bucket.s3Client.CopyObject(&s3.CopyObjectInput{
		Bucket:     aws.String(bucket.Name()),
		Key:        aws.String(destinationKey),
		CopySource: aws.String(copySourceString),
	})
	return err
}

type partUploadOutput struct {
	completedPart *s3.CompletedPart
	err           error
}

func (bucket Bucket) copyVersionWithMultipart(sourceBucketName, blobKey, copySourceString, destinationPath string, blobSize int64) error {
	var destinationKey string
	if destinationPath != "" {
		destinationKey = fmt.Sprintf("%s/%s", destinationPath, blobKey)
	} else {
		destinationKey = blobKey
	}

	createOutput, err := bucket.s3Client.CreateMultipartUpload(
		&s3.CreateMultipartUploadInput{
			Bucket: aws.String(bucket.Name()),
			Key:    aws.String(destinationKey),
		},
	)

	if err != nil {
		return fmt.Errorf("failed to create multipart upload: %s", err)
	}

	numParts := int64(math.Ceil(float64(blobSize) / float64(partSize)))
	var parts []*s3.CompletedPart
	var errors []error

	for partNumber := int64(1); partNumber <= numParts; partNumber++ {
		partStart := (partNumber - 1) * partSize
		partEnd := int64(math.Min(float64(partNumber*partSize-1), float64(blobSize-1)))

		copyPartOutput, err := bucket.s3Client.UploadPartCopy(&s3.UploadPartCopyInput{
			Bucket:          aws.String(bucket.Name()),
			Key:             aws.String(destinationKey),
			UploadId:        aws.String(*createOutput.UploadId),
			CopySource:      aws.String(copySourceString),
			CopySourceRange: aws.String(fmt.Sprintf("bytes=%d-%d", partStart, partEnd)),
			PartNumber:      aws.Int64(partNumber),
		})

		if err != nil {
			errors = append(errors, fmt.Errorf("failed to upload part with range: %d-%d: %s", partStart, partEnd, err))
		} else {
			parts = append(parts, &s3.CompletedPart{
				PartNumber: aws.Int64(partNumber),
				ETag:       copyPartOutput.CopyPartResult.ETag,
			})
		}
	}

	if len(errors) != 0 {
		_, err := bucket.s3Client.AbortMultipartUpload(&s3.AbortMultipartUploadInput{
			Bucket:   aws.String(bucket.Name()),
			Key:      aws.String(destinationKey),
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
		Bucket:   aws.String(bucket.Name()),
		Key:      aws.String(destinationKey),
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

func formatErrors(contextString string, errors []error) error {
	errorStrings := make([]string, len(errors))
	for i, err := range errors {
		errorStrings[i] = err.Error()
	}
	return fmt.Errorf("%s: %s", contextString, strings.Join(errorStrings, "\n"))
}

func newS3Client(regionName string, endpoint string, accessKey AccessKey, useIAMProfile bool) (*s3.S3, error) {
	var creds = credentials.NewStaticCredentials(accessKey.Id, accessKey.Secret, "")

	if useIAMProfile {
		s, err := session.NewSession(aws.NewConfig().WithRegion(regionName))
		if err != nil {
			return nil, err
		}

		creds = ec2rolecreds.NewCredentials(s)
	}

	awsSession, err := session.NewSession(&aws.Config{
		Region:           &regionName,
		Credentials:      creds,
		Endpoint:         aws.String(endpoint),
		S3ForcePathStyle: aws.Bool(true),
	})

	if err != nil {
		return nil, err
	}

	return s3.New(awsSession), nil
}
