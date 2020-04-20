package s3bucket

import (
	"fmt"
	"math"
	"sort"

	"s3-blobstore-backup-restore/blobpath"

	"github.com/aws/aws-sdk-go/aws/awserr"

	"s3-blobstore-backup-restore/incremental"

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

func NewBucket(bucketName, bucketRegion, endpoint string, accessKey AccessKey, useIAMProfile, forcePathStyle bool) (Bucket, error) {
	s3Client, err := newS3Client(bucketRegion, endpoint, accessKey, useIAMProfile, forcePathStyle)
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

func (b Bucket) listFiles(subfolder string) ([]string, error) {
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
			files[index] = strings.Replace(fileLocation, subfolder+blobpath.Delimiter, "", 1)
		}
	}

	return files, nil
}

func (b Bucket) ListBlobs(prefix string) ([]incremental.Blob, error) {
	paths, err := b.listFiles(prefix)
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
		Delimiter: aws.String(blobpath.Delimiter),
	}, func(output *s3.ListObjectsOutput, lastPage bool) bool {
		for _, value := range output.CommonPrefixes {
			dirs = append(dirs, blobpath.TrimTrailingDelimiter(*value.Prefix))
		}

		return !lastPage
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list directories from bucket %s: %s", b.name, err)
	}

	return dirs, nil
}

func (b Bucket) ListVersions() ([]Version, error) {
	isVersioned, err := b.IsVersioned()
	if err != nil {
		return nil, err
	}

	if !isVersioned {
		return nil, fmt.Errorf("bucket %s is not versioned", b.Name())
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

func (b Bucket) IsVersioned() (bool, error) {
	output, err := b.s3Client.GetBucketVersioning(&s3.GetBucketVersioningInput{
		Bucket: &b.name,
	})

	if err != nil {
		return false, fmt.Errorf("could not check if bucket %s is versioned: %s", b.name, err)
	}

	if output == nil || output.Status == nil || *output.Status != "Enabled" {
		return false, nil
	}

	return true, nil
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

	copySource := blobpath.Delimiter + originBucketName + blobpath.Delimiter + blobKey
	if versionID != "null" {
		copySource = copySource + "?versionId=" + versionID
	}

	copySource = strings.Replace(copySource, blobpath.Delimiter+blobpath.Delimiter, blobpath.Delimiter, -1)

	if sizeInMbs(blobSize) <= 1024 {
		return b.copyVersionWithSingleRequest(copySource, destinationKey)
	} else {
		return b.copyVersionWithMultipart(copySource, destinationKey, blobSize)
	}
}

var injectableNewS3Client = newS3Client

func (b Bucket) getBlobSize(bucketName, bucketRegion, blobKey, versionID string) (int64, error) {
	s3Client, err := injectableNewS3Client(bucketRegion, b.endpoint, b.accessKey, b.useIAMProfile, *b.s3Client.Client.Config.S3ForcePathStyle)
	if err != nil {
		return 0, err
	}

	input := s3.HeadObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(blobKey),
	}
	if versionID != "null" {
		input.SetVersionId(versionID)
	}

	headObjectOutput, err := s3Client.HeadObject(&input)

	if err != nil {
		return 0, fmt.Errorf("failed to get blob size for blob '%s' in bucket '%s': %s", blobKey, bucketName, err)
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

var injectableCredIAMProvider = ec2rolecreds.NewCredentials

func newS3Client(regionName, endpoint string, accessKey AccessKey, useIAMProfile, forcePathStyle bool) (*s3.S3, error) {
	var creds = credentials.NewStaticCredentials(accessKey.Id, accessKey.Secret, "")

	if useIAMProfile {
		s, err := session.NewSession(aws.NewConfig().WithRegion(regionName))
		if err != nil {
			return nil, err
		}

		creds = injectableCredIAMProvider(s)
	}

	awsSession, err := session.NewSession(&aws.Config{
		Region:           &regionName,
		Credentials:      creds,
		Endpoint:         aws.String(endpoint),
		S3ForcePathStyle: aws.Bool(forcePathStyle),
	})

	if err != nil {
		return nil, err
	}

	return s3.New(awsSession), nil
}
