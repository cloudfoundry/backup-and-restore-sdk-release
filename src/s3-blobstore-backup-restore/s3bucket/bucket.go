package s3bucket

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"

	"s3-blobstore-backup-restore/blobpath"
	"s3-blobstore-backup-restore/incremental"
)

const partSize int64 = 100 * 1024 * 1024

type Bucket struct {
	name           string
	regionName     string
	accessKey      AccessKey
	endpoint       string
	s3Client       *s3.Client
	useIAMProfile  bool
	forcePathStyle bool
	assumedRoleARN string
	clientOptFns   []func(*s3.Options)
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

func NewBucket(bucketName, bucketRegion, endpoint string, accessKey AccessKey, useIAMProfile, forcePathStyle bool, clientOptFns ...func(*s3.Options)) (Bucket, error) {
	s3Client, err := newS3Client(bucketRegion, endpoint, accessKey, useIAMProfile, forcePathStyle, clientOptFns...)
	if err != nil {
		return Bucket{}, err
	}

	return Bucket{
		name:           bucketName,
		regionName:     bucketRegion,
		s3Client:       s3Client,
		accessKey:      accessKey,
		endpoint:       endpoint,
		useIAMProfile:  useIAMProfile,
		forcePathStyle: forcePathStyle,
		clientOptFns:   clientOptFns,
	}, nil
}

// NewBucketWithRoleARN
//
// # Warning!
//
// Utilising the assumed role is a highly experimental functionality and is provided as is. Use at your own risk.
func NewBucketWithRoleARN(bucketName, bucketRegion, endpoint, roleARN string, accessKey AccessKey, useIAMProfile, forcePathStyle bool, clientOptFns ...func(*s3.Options)) (Bucket, error) {
	s3Client, err := newS3ClientWithAssumedRole(bucketRegion, endpoint, accessKey, useIAMProfile, forcePathStyle, roleARN, clientOptFns...)
	if err != nil {
		return Bucket{}, err
	}

	return Bucket{
		name:           bucketName,
		regionName:     bucketRegion,
		s3Client:       s3Client,
		accessKey:      accessKey,
		endpoint:       endpoint,
		useIAMProfile:  useIAMProfile,
		forcePathStyle: forcePathStyle,
		assumedRoleARN: roleARN,
		clientOptFns:   clientOptFns,
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

	params := &s3.ListObjectsV2Input{
		Bucket: aws.String(b.name),
		Prefix: aws.String(subfolder),
	}

	paginator := s3.NewListObjectsV2Paginator(b.s3Client, params)
	for paginator.HasMorePages() {
		output, err := paginator.NextPage(context.TODO())
		if err != nil {
			return nil, fmt.Errorf("failed to list blobs from bucket %s: %s", b.name, err)
		}
		for _, file := range output.Contents {
			files = append(files, *file.Key)
		}
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
	params := &s3.ListObjectsV2Input{
		Bucket:    aws.String(b.name),
		Prefix:    aws.String(""),
		Delimiter: aws.String(blobpath.Delimiter),
	}

	paginator := s3.NewListObjectsV2Paginator(b.s3Client, params)
	for paginator.HasMorePages() {
		output, err := paginator.NextPage(context.TODO())
		if err != nil {
			return nil, fmt.Errorf("failed to list directories from bucket %s: %s", b.name, err)
		}
		for _, value := range output.CommonPrefixes {
			dirs = append(dirs, blobpath.TrimTrailingDelimiter(*value.Prefix))
		}
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

	paginator := s3.NewListObjectVersionsPaginator(b.s3Client, &s3.ListObjectVersionsInput{
		Bucket:  aws.String(b.name),
		MaxKeys: aws.Int32(1000),
	})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(context.TODO())
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve versions from b %s: %s", b.name, err)
		}

		for _, v := range output.Versions {
			version := Version{
				Key:      *v.Key,
				Id:       *v.VersionId,
				IsLatest: *v.IsLatest,
			}
			versions = append(versions, version)
		}
	}

	return versions, nil
}

func (b Bucket) IsVersioned() (bool, error) {
	output, err := b.s3Client.GetBucketVersioning(context.TODO(), &s3.GetBucketVersioningInput{
		Bucket: &b.name,
	})

	if err != nil {
		return false, fmt.Errorf("could not check if bucket %s is versioned: %s", b.name, err)
	}

	if output == nil || output.Status != types.BucketVersioningStatusEnabled {
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
	_, err := b.s3Client.PutObject(context.TODO(), &s3.PutObjectInput{
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
	_, err := b.s3Client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(b.name),
		Key:    aws.String(key),
	})
	if err != nil {
		var nsk *types.NoSuchKey
		if errors.As(err, &nsk) {
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

func (b Bucket) getBlobSize(bucketName, bucketRegion, blobKey, versionID string) (int64, error) {
	s3Client, err := newS3ClientWithAssumedRole(bucketRegion, b.endpoint, b.accessKey, b.useIAMProfile, b.forcePathStyle, b.assumedRoleARN, b.clientOptFns...)
	if err != nil {
		return 0, err
	}

	input := s3.HeadObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(blobKey),
	}
	if versionID != "null" {
		input.VersionId = &versionID
	}

	headObjectOutput, err := s3Client.HeadObject(context.TODO(), &input)

	if err != nil {
		return 0, fmt.Errorf("failed to get blob size for blob '%s' in bucket '%s': %s", blobKey, bucketName, err)
	}

	return *headObjectOutput.ContentLength, nil
}

func sizeInMbs(sizeInBytes int64) int64 {
	return sizeInBytes / (1024 * 1024)
}

func (b Bucket) copyVersionWithSingleRequest(copySourceString, destinationKey string) error {
	_, err := b.s3Client.CopyObject(context.TODO(), &s3.CopyObjectInput{
		Bucket:     aws.String(b.name),
		Key:        aws.String(destinationKey),
		CopySource: aws.String(copySourceString),
	})
	return err
}

func (b Bucket) copyVersionWithMultipart(copySourceString, destinationKey string, blobSize int64) error {
	createOutput, err := b.s3Client.CreateMultipartUpload(context.TODO(),
		&s3.CreateMultipartUploadInput{
			Bucket: aws.String(b.Name()),
			Key:    aws.String(destinationKey),
		},
	)

	if err != nil {
		return fmt.Errorf("failed to create multipart upload: %s", err)
	}

	numParts := int32(math.Ceil(float64(blobSize) / float64(partSize)))
	var parts []types.CompletedPart
	var uploadErrors []error
	var partNumber int32

	for partNumber = 1; partNumber <= numParts; partNumber++ {
		partStart := int64(partNumber-1) * partSize
		partEnd := int64(math.Min(float64(int64(partNumber)*partSize-1), float64(blobSize-1)))

		copyPartOutput, err := b.s3Client.UploadPartCopy(context.TODO(), &s3.UploadPartCopyInput{
			Bucket:          aws.String(b.Name()),
			Key:             aws.String(destinationKey),
			UploadId:        aws.String(*createOutput.UploadId),
			CopySource:      aws.String(copySourceString),
			CopySourceRange: aws.String(fmt.Sprintf("bytes=%d-%d", partStart, partEnd)),
			PartNumber:      aws.Int32(partNumber),
		})

		if err != nil {
			uploadErrors = append(uploadErrors, fmt.Errorf("failed to upload part with range: %d-%d: %s", partStart, partEnd, err))
		} else {
			parts = append(parts, types.CompletedPart{
				PartNumber: aws.Int32(partNumber),
				ETag:       copyPartOutput.CopyPartResult.ETag,
			})
		}
	}

	if len(uploadErrors) != 0 {
		_, err := b.s3Client.AbortMultipartUpload(context.TODO(), &s3.AbortMultipartUploadInput{
			Bucket:   aws.String(b.Name()),
			Key:      aws.String(destinationKey),
			UploadId: createOutput.UploadId,
		})

		if err != nil {
			uploadErrors = append(uploadErrors, err)
		}

		return formatErrors("errors occurred in multipart upload", uploadErrors)
	}

	sort.Slice(parts, func(i, j int) bool {
		return *parts[i].PartNumber < *parts[j].PartNumber
	})

	_, err = b.s3Client.CompleteMultipartUpload(context.TODO(), &s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(b.Name()),
		Key:      aws.String(destinationKey),
		UploadId: createOutput.UploadId,
		MultipartUpload: &types.CompletedMultipartUpload{
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

func newS3Client(regionName, endpoint string, accessKey AccessKey, useIAMProfile, forcePathStyle bool, fns ...func(*s3.Options)) (*s3.Client, error) {
	return newS3ClientWithAssumedRole(regionName, endpoint, accessKey, useIAMProfile, forcePathStyle, "", fns...)
}

func newS3ClientWithAssumedRole(regionName, endpoint string, accessKey AccessKey, useIAMProfile, forcePathStyle bool, role string, fns ...func(*s3.Options)) (*s3.Client, error) {
	staticCredentialsProvider := aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(accessKey.Id, accessKey.Secret, ""))

	var creds aws.CredentialsProvider

	if role != "" {
		stsOptions := sts.Options{
			Credentials: staticCredentialsProvider,
			Region:      regionName,
		}
		if endpoint != "" {
			stsOptions.EndpointResolver = sts.EndpointResolverFromURL(endpoint)
		}
		stsClient := sts.New(stsOptions)
		creds = stscreds.NewAssumeRoleProvider(stsClient, role)
	} else if useIAMProfile {
		creds = aws.NewCredentialsCache(ec2rolecreds.New())
	} else {
		creds = staticCredentialsProvider
	}

	options := s3.Options{
		Credentials:  creds,
		Region:       regionName,
		UsePathStyle: forcePathStyle,
	}

	if endpoint != "" {
		options.EndpointResolver = s3.EndpointResolverFromURL(endpoint)
	}

	client := s3.New(options, fns...)

	return client, nil
}
