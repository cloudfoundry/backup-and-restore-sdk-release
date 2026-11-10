module azure-blobstore-backup-restore

go 1.26.11

require (
	github.com/Azure/azure-sdk-for-go/sdk/storage/azblob v1.8.3
	github.com/maxbrunsfeld/counterfeiter/v6 v6.14.0
	github.com/onsi/ginkgo/v2 v2.34.0
	github.com/onsi/gomega v1.44.0
	system-tests v0.0.0
)

require (
	github.com/Azure/azure-sdk-for-go/sdk/azcore v1.24.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.15.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/internal v1.13.0 // indirect
	github.com/Azure/azure-sdk-for-go/sdk/storage/internal v0.1.0 // indirect
	github.com/Masterminds/semver/v3 v3.5.0 // indirect
	github.com/andybalholm/brotli v1.3.0 // indirect
	github.com/apache/arrow-go/v18 v18.8.0 // indirect
	github.com/apache/thrift v0.25.0 // indirect
	github.com/gkampitakis/ciinfo v0.3.4 // indirect
	github.com/go-logr/logr v1.4.4 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/goccy/go-json v0.12.0 // indirect
	github.com/goccy/go-yaml v1.19.2 // indirect
	github.com/google/flatbuffers v25.12.19+incompatible // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/pprof v0.0.0-20261110164715-8bb1ee682bd8 // indirect
	github.com/klauspost/compress v1.20.4 // indirect
	github.com/klauspost/cpuid/v2 v2.4.2 // indirect
	github.com/maruel/natural v1.3.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.33 // indirect
	github.com/sergi/go-diff v1.4.0 // indirect
	github.com/zeebo/assert v1.3.1 // indirect
	github.com/zeebo/xxh3 v1.1.0 // indirect
	go.yaml.in/yaml/v3 v3.0.6 // indirect
	golang.org/x/exp v0.0.0-20261109185158-a52e3aef5920 // indirect
	golang.org/x/mod v0.43.0 // indirect
	golang.org/x/net v0.62.0 // indirect
	golang.org/x/sync v0.25.0 // indirect
	golang.org/x/sys v0.50.0 // indirect
	golang.org/x/text v0.44.0 // indirect
	golang.org/x/tools v0.52.1-0.20261106145633-d407941a339a // indirect
)

replace system-tests => ../system-tests

replace s3-blobstore-backup-restore => ../s3-blobstore-backup-restore
