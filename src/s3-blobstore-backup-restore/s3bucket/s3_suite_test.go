package s3bucket_test

import (
	"fmt"
	"os"
	"path"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	AssumedRoleProfileName   = "assumed-role-profile"
	awsProfileConfigTemplate = `[profile %s]
    role_arn = %s
    credential_source = Environment`
)

var (
	S3Endpoint                   = MustHaveEnvOrBeEmpty("S3_ENDPOINT")
	LiveRegion                   = MustHaveEnv("S3_LIVE_REGION")
	BackupRegion                 = MustHaveEnv("S3_BACKUP_REGION")
	AccessKey                    = MustHaveEnv("S3_ACCESS_KEY_ID")
	SecretKey                    = MustHaveEnv("S3_SECRET_ACCESS_KEY")
	AssumedRoleARN               = MustHaveEnv("S3_ASSUMED_ROLE_ARN")
	PreExistingBigFileBucketName = MustHaveEnv("S3_BIG_FILE_BUCKET")
)

func TestS3(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "S3 Suite")
}

var _ = BeforeSuite(func() {
	mustCreateAWSConfigFile()
})

func mustCreateAWSConfigFile() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic("Unable to fetch the user home directory path")
	}
	err = os.Mkdir(path.Join(homeDir, ".aws"), 0755)
	if err != nil {
		panic("Unable to create AWS settings directory")
	}

	configFile, err := os.Create(path.Join(homeDir, ".aws", "config"))
	if err != nil {
		panic("Unable to create AWS settings file")
	}
	defer configFile.Close()

	_, err = fmt.Fprintf(configFile, awsProfileConfigTemplate, AssumedRoleProfileName, AssumedRoleARN)
	if err != nil {
		panic("Unable to write AWS settings file")
	}
}

func MustHaveEnvOrBeEmpty(keyName string) string {
	val, exist := os.LookupEnv(keyName)
	if !exist {
		panic("Need " + keyName + " for the test")
	}
	return val
}

func MustHaveEnv(keyName string) string {
	val := os.Getenv(keyName)
	if val == "" {
		panic("Need " + keyName + " for the test")
	}
	return val
}
