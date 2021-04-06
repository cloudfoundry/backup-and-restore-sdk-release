module s3-blobstore-backup-restore

go 1.14

require (
	github.com/aws/aws-sdk-go v1.38.12
	github.com/cloudfoundry-incubator/bosh-backup-and-restore v1.9.6
	github.com/onsi/ginkgo v1.16.0
	github.com/onsi/gomega v1.11.0
	system-tests v0.0.0
)

replace system-tests => ../system-tests

replace s3-blobstore-backup-restore => ./
