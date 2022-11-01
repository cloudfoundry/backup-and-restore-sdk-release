module system-tests

go 1.18

require (
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.19.0
	s3-blobstore-backup-restore v0.0.0
)

replace s3-blobstore-backup-restore => ../s3-blobstore-backup-restore
