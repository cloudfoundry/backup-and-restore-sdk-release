module s3-blobstore-backup-restore

require (
	github.com/aws/aws-sdk-go v1.20.15
	github.com/cloudfoundry-incubator/bosh-backup-and-restore v1.5.1
	github.com/onsi/ginkgo v1.8.0
	github.com/onsi/gomega v1.5.0
	system-tests v0.0.0
)

replace system-tests => ../system-tests

replace s3-blobstore-backup-restore => ./
