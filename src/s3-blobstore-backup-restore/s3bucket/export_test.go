package s3bucket

import (
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
)

var NewS3ClientImpl = newS3Client

type CredIAMProvider = func(c client.ConfigProvider, options ...func(*ec2rolecreds.EC2RoleProvider)) *credentials.Credentials

func SetCredIAMProvider(provider CredIAMProvider) {
	injectableCredIAMProvider = provider
}

type NewS3Client = func(regionName, endpoint string, accessKey AccessKey, useIAMProfile, forcePathStyle bool) (*s3.S3, error)

func SetNewS3Client(s3Client NewS3Client) {
	injectableNewS3Client = s3Client
}

func (b Bucket) UsesPathStyle() bool {
	return *b.s3Client.Client.Config.S3ForcePathStyle
}

func (b Bucket) GetBlobSizeImpl(bucketName, bucketRegion, blobKey, versionID string) (int64, error) {
	return b.getBlobSize(bucketName, bucketRegion, blobKey, versionID)
}
