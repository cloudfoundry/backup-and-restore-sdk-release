module azure-blobstore-backup-restore

go 1.23.0

require (
	github.com/Azure/azure-sdk-for-go/sdk/storage/azblob v1.6.3
	github.com/maxbrunsfeld/counterfeiter/v6 v6.11.3
	github.com/onsi/ginkgo/v2 v2.26.0
	github.com/onsi/gomega v1.38.2
	system-tests v0.0.0
)

require (
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.19.1 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.11.2 // indirect
	github.com/Masterminds/semver/v3 v3.4.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/pprof v0.0.0-20250630185457-6e76a2b096b5 // indirect
	go.uber.org/automaxprocs v1.6.0 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/mod v0.27.0 // indirect
	golang.org/x/net v0.43.0 // indirect
	golang.org/x/sync v0.16.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
	golang.org/x/text v0.28.0 // indirect
	golang.org/x/tools v0.36.0 // indirect
)

replace system-tests => ../system-tests

replace s3-blobstore-backup-restore => ../s3-blobstore-backup-restore
