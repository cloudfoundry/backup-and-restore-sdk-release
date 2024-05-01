module azure-blobstore-backup-restore

go 1.22.2

require (
	github.com/Azure/azure-sdk-for-go/sdk/storage/azblob v1.3.2
	github.com/maxbrunsfeld/counterfeiter/v6 v6.8.1
	github.com/onsi/ginkgo/v2 v2.17.2
	github.com/onsi/gomega v1.33.1
	system-tests v0.0.0
)

require (
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.11.1 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.5.2 // indirect
	github.com/go-logr/logr v1.4.1 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/google/pprof v0.0.0-20240424215950-a892ee059fd6 // indirect
	github.com/kr/text v0.1.0 // indirect
	golang.org/x/mod v0.17.0 // indirect
	golang.org/x/net v0.24.0 // indirect
	golang.org/x/sync v0.7.0 // indirect
	golang.org/x/sys v0.19.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	golang.org/x/tools v0.20.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace system-tests => ../system-tests

replace s3-blobstore-backup-restore => ../s3-blobstore-backup-restore
