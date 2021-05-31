module azure-blobstore-backup-restore

go 1.16

require (
	github.com/Azure/azure-pipeline-go v0.2.1 // indirect
	github.com/Azure/azure-storage-blob-go v0.0.0-20180507052152-66ba96e49ebb
	github.com/onsi/ginkgo v1.16.3
	github.com/onsi/gomega v1.11.0
	system-tests v0.0.0
)

replace system-tests => ../system-tests

replace azure-blobstore-backup-restore => ./

replace s3-blobstore-backup-restore => ../s3-blobstore-backup-restore
