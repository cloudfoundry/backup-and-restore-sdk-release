package s3_test

import (
	. "github.com/cloudfoundry-incubator/backup-and-restore-sdk-release-system-tests"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

var (
	TestAWSAccessKeyID     = MustHaveEnv("TEST_AWS_ACCESS_KEY_ID")
	TestAWSSecretAccessKey = MustHaveEnv("TEST_AWS_SECRET_ACCESS_KEY")

	TestECSAccessKeyID     = MustHaveEnv("TEST_ECS_ACCESS_KEY_ID")
	TestECSSecretAccessKey = MustHaveEnv("TEST_ECS_SECRET_ACCESS_KEY")
)

func TestS3(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "S3 Suite")
}
