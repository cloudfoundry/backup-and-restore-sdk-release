package s3

import (
	"fmt"
	"math"
	"sort"

	"github.com/aws/aws-sdk-go/aws/awserr"

	"github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/incremental"

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
	Region() string
	CopyObject(blobKey, originPath, destinationPath, originBucketName, originBucketRegion string) error
	ListFiles(path string) ([]string, error)
}

//go:generate counterfeiter -o fakes/fake_versioned_bucket.go . VersionedBucket
type VersionedBucket interface {
	Name() string
	Region() string
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

func (b Bucket) Name() string {
	return b.name
}

func (b Bucket) Region() string {
	return b.regionName
}

func (b Bucket) ListFiles(subfolder string) ([]string, error) {
	var files []string
	err := b.s3Client.ListObjectsPages(&s3.ListObjectsInput{
		Bucket: aws.String(b.name),
		Prefix: aws.String(subfolder),
	}, func(output *s3.ListObjectsOutput, lastPage bool) bool {
		for _, value := range output.Contents {
			files = append(files, *value.Key)
		}

		return !lastPage
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list blobs from bucket %s: %s", b.name, err)
	}

	if subfolder != "" {
		for index, fileLocation := range files {
			files[index] = strings.Replace(fileLocation, subfolder+blobDelimiter, "", 1)
		}
	}

	return files, nil
}

func (b Bucket) ListBlobs(prefix string) ([]incremental.Blob, error) {
	paths, err := b.ListFiles(prefix)
	if err != nil {
		return nil, err
	}

	var blobs []incremental.Blob
	for _, path := range paths {
		blobs = append(blobs, NewBlob(path))
	}

	return blobs, err
}

func (b Bucket) ListDirectories() ([]string, error) {
	var dirs []string
	err := b.s3Client.ListObjectsPages(&s3.ListObjectsInput{
		Bucket:    aws.String(b.name),
		Prefix:    aws.String(""),
		Delimiter: aws.String(blobDelimiter),
	}, func(output *s3.ListObjectsOutput, lastPage bool) bool {
		for _, value := range output.CommonPrefixes {
			dirs = append(dirs, strings.TrimSuffix(*value.Prefix, blobDelimiter))
		}

		return !lastPage
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list directories from bucket %s: %s", b.name, err)
	}

	return dirs, nil
}

func (b Bucket) ListVersions() ([]Version, error) {
	err := b.CheckIfVersioned()
	if err != nil {
		return nil, err
	}

	var versions []Version
	err = b.s3Client.ListObjectVersionsPages(&s3.ListObjectVersionsInput{
		Bucket: aws.String(b.name),
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
		return nil, fmt.Errorf("failed to retrieve versions from b %s: %s", b.name, err)
	}

	return versions, nil
}

func (b Bucket) CheckIfVersioned() error {
	output, err := b.s3Client.GetBucketVersioning(&s3.GetBucketVersioningInput{
		Bucket: &b.name,
	})

	if err != nil {
		return fmt.Errorf("could not check if bucket %s is versioned: %s", b.name, err)
	}

	if output == nil || output.Status == nil || *output.Status != "Enabled" {
		return fmt.Errorf("b %s is not versioned", b.name)
	}

	return nil
}

func (b Bucket) CopyBlobWithinBucket(src, dst string) error {
	return b.copyVersion(src, "null", dst, b.name, b.regionName)
}

func (b Bucket) CopyBlobFromBucket(sourceBucket incremental.Bucket, src, dst string) error {
	srcBucket := sourceBucket.(Bucket)
	return b.copyVersion(src, "null", dst, srcBucket.name, srcBucket.regionName)
}

func (b Bucket) UploadBlob(key, contents string) error {
	_, err := b.s3Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(b.name),
		Key:    aws.String(key),
		Body:   strings.NewReader(contents),
	})
	if err != nil {
		return fmt.Errorf("failed to upload blob '%s': %s", key, err)
	}

	return nil
}

func (b Bucket) HasBlob(key string) (bool, error) {
	_, err := b.s3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(b.name),
		Key:    aws.String(key),
	})
	if err != nil {
		awsErr, ok := err.(awserr.Error)
		if ok && awsErr.Code() == s3.ErrCodeNoSuchKey {
			return false, nil
		}
		return false, fmt.Errorf("failed to check if blob exists '%s': %s", key, err)
	}

	return true, nil
}

func (b Bucket) CopyObject(originSuffix, originPrefix, destinationPrefix, originBucketName, originBucketRegion string) error {
	blobKey := strings.Join([]string{originPrefix, originSuffix}, blobDelimiter)
	destinationKey := strings.Join([]string{destinationPrefix, originSuffix}, blobDelimiter)

	return b.copyVersion(
		blobKey,
		"null",
		destinationKey,
		originBucketName,
		originBucketRegion,
	)
}

func (b Bucket) CopyVersion(blobKey, versionID, originBucketName, originBucketRegion string) error {
	return b.copyVersion(
		blobKey,
		versionID,
		blobKey,
		originBucketName,
		originBucketRegion,
	)
}

func (b Bucket) copyVersion(blobKey, versionID, destinationKey, originBucketName, originBucketRegion string) error {
	blobSize, err := b.getBlobSize(originBucketName, originBucketRegion, blobKey, versionID)
	if err != nil {
		return err
	}

	copySource := blobDelimiter + originBucketName + blobDelimiter + blobKey + "?versionId=" + versionID
	copySource = strings.Replace(copySource, blobDelimiter+blobDelimiter, blobDelimiter, -1)

	if sizeInMbs(blobSize) <= 1024 {
		return b.copyVersionWithSingleRequest(copySource, destinationKey)
	} else {
		return b.copyVersionWithMultipart(copySource, destinationKey, blobSize)
	}
}

func (b Bucket) getBlobSize(bucketName, bucketRegion, blobKey, versionID string) (int64, error) {
	s3Client, err := newS3Client(bucketRegion, b.endpoint, b.accessKey, b.useIAMProfile)
	if err != nil {
		return 0, err
	}

	headObjectOutput, err := s3Client.HeadObject(&s3.HeadObjectInput{
		Bucket:    aws.String(bucketName),
		Key:       aws.String(blobKey),
		VersionId: aws.String(versionID),
	})

	if err != nil {
		return 0, fmt.Errorf("failed to retrieve head object output: %s", err)
	}

	return *headObjectOutput.ContentLength, nil
}

func sizeInMbs(sizeInBytes int64) int64 {
	return sizeInBytes / (1024 * 1024)
}

func (b Bucket) copyVersionWithSingleRequest(copySourceString, destinationKey string) error {
	_, err := b.s3Client.CopyObject(&s3.CopyObjectInput{
		Bucket:     aws.String(b.name),
		Key:        aws.String(destinationKey),
		CopySource: aws.String(copySourceString),
	})
	return err
}

type partUploadOutput struct {
	completedPart *s3.CompletedPart
	err           error
}

func (b Bucket) copyVersionWithMultipart(copySourceString, destinationKey string, blobSize int64) error {
	createOutput, err := b.s3Client.CreateMultipartUpload(
		&s3.CreateMultipartUploadInput{
			Bucket: aws.String(b.Name()),
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

		copyPartOutput, err := b.s3Client.UploadPartCopy(&s3.UploadPartCopyInput{
			Bucket:          aws.String(b.Name()),
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
		_, err := b.s3Client.AbortMultipartUpload(&s3.AbortMultipartUploadInput{
			Bucket:   aws.String(b.Name()),
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

	_, err = b.s3Client.CompleteMultipartUpload(&s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(b.Name()),
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
