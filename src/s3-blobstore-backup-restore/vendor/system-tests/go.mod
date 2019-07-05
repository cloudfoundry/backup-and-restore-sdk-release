module system-tests

require (
	github.com/onsi/ginkgo v1.8.0
	github.com/onsi/gomega v1.5.0
	s3-blobstore-backup-restore v0.0.0
)

replace system-tests => ./

replace s3-blobstore-backup-restore => ../s3-blobstore-backup-restore
