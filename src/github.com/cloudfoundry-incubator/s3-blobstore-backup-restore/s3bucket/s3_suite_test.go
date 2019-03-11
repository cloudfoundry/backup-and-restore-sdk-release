package s3bucket_test

import (
	. "github.com/cloudfoundry-incubator/backup-and-restore-sdk-release-system-tests"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

var (
	LiveRegion                   = MustHaveEnv("S3_LIVE_REGION")
	BackupRegion                 = MustHaveEnv("S3_BACKUP_REGION")
	S3Endpoint                   = MustHaveEnv("S3_ENDPOINT")
	AccessKey                    = MustHaveEnv("S3_ACCESS_KEY_ID")
	SecretKey                    = MustHaveEnv("S3_SECRET_ACCESS_KEY")
	PreExistingBigFileBucketName = MustHaveEnv("S3_BIG_FILE_BUCKET")
	EmptyBucketName              = MustHaveEnv("S3_EMPTY_BUCKET")
)

func TestS3(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "S3 Suite")
}
