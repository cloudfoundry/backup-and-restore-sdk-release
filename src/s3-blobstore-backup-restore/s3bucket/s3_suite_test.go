package s3bucket_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"os"
	"testing"
)

var (
	S3Endpoint                   = MustHaveEnvOrBeEmpty("S3_ENDPOINT")
	LiveRegion                   = MustHaveEnv("S3_LIVE_REGION")
	BackupRegion                 = MustHaveEnv("S3_BACKUP_REGION")
	AccessKey                    = MustHaveEnv("S3_ACCESS_KEY_ID")
	SecretKey                    = MustHaveEnv("S3_SECRET_ACCESS_KEY")
	PreExistingBigFileBucketName = MustHaveEnv("S3_BIG_FILE_BUCKET")
)

func TestS3(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "S3 Suite")
}

func MustHaveEnvOrBeEmpty(keyname string) string {
	val, exist := os.LookupEnv(keyname)
	if !exist {
		panic("Need " + keyname + " for the test")
	}
	return val
}

func MustHaveEnv(keyname string) string {
	val := os.Getenv(keyname)
	if val == "" {
		panic("Need " + keyname + " for the test")
	}
	return val
}
