module azure-blobstore-backup-restore

go 1.16

require (
	github.com/Azure/azure-storage-blob-go v0.15.0
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.23.0
	system-tests v0.0.0
)

replace system-tests => ../system-tests

replace s3-blobstore-backup-restore => ../s3-blobstore-backup-restore
