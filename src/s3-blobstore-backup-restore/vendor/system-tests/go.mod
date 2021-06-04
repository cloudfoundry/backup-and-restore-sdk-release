module system-tests

go 1.16

require (
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.13.0
	s3-blobstore-backup-restore v0.0.0
)

replace system-tests => ./

replace s3-blobstore-backup-restore => ../s3-blobstore-backup-restore
