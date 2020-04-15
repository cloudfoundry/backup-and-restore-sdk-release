package s3bucket

import (
	"github.com/aws/aws-sdk-go/aws/client"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
)

var NewS3ClientImpl = newS3Client
type CredIAMProvider = func(c client.ConfigProvider, options ...func(*ec2rolecreds.EC2RoleProvider)) *credentials.Credentials 

func SetCredIAMProvider(provider CredIAMProvider) {
	injectableCredIAMProvider = provider
}

func (b Bucket) UsesPathStyle() bool {
	return *b.s3Client.Client.Config.S3ForcePathStyle
}
