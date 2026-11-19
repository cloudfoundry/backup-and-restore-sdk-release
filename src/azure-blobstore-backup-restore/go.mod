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
	github.com/apache/arrow-go/v18 v18.9.0 // indirect
	github.com/go-logr/logr v1.4.4 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/goccy/go-json v0.12.0 // indirect
	github.com/goccy/go-yaml v1.19.2 // indirect
	github.com/google/flatbuffers v25.12.19+incompatible // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/pprof v0.0.0-20261118010117-084b33a85f63 // indirect
	github.com/klauspost/compress v1.20.4 // indirect
	github.com/klauspost/cpuid/v2 v2.4.2 // indirect
	github.com/maruel/natural v1.3.0 // indirect
	github.com/pierrec/lz4/v4 v4.1.33 // indirect
	github.com/sergi/go-diff v1.4.5 // indirect
	github.com/tidwall/gjson v1.20.4 // indirect
	github.com/tidwall/sjson v1.2.10 // indirect
	github.com/zeebo/assert v1.3.1 // indirect
	github.com/zeebo/xxh3 v1.1.0 // indirect
	go.yaml.in/yaml/v3 v3.0.6 // indirect
	golang.org/x/exp v0.0.0-20261119141654-b15d094f3610 // indirect
	golang.org/x/mod v0.43.0 // indirect
	golang.org/x/net v0.63.0 // indirect
	golang.org/x/sync v0.25.0 // indirect
	golang.org/x/sys v0.50.0 // indirect
	golang.org/x/text v0.44.0 // indirect
	golang.org/x/tools v0.53.0 // indirect
)

replace system-tests => ../system-tests

replace s3-blobstore-backup-restore => ../s3-blobstore-backup-restore
