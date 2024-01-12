module azure-blobstore-backup-restore

go 1.21

toolchain go1.21.0

require (
	github.com/Azure/azure-storage-blob-go v0.15.0
	github.com/maxbrunsfeld/counterfeiter/v6 v6.7.0
	github.com/onsi/ginkgo/v2 v2.14.0
	github.com/onsi/gomega v1.30.0
	system-tests v0.0.0
)

require (
	github.com/Azure/azure-pipeline-go v0.2.3 // indirect
	github.com/go-logr/logr v1.3.0 // indirect
	github.com/go-task/slim-sprig v0.0.0-20230315185526-52ccab3ef572 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/google/pprof v0.0.0-20230323073829-e72429f035bd // indirect
	github.com/google/uuid v1.2.0 // indirect
	github.com/mattn/go-ieproxy v0.0.1 // indirect
	golang.org/x/mod v0.14.0 // indirect
	golang.org/x/net v0.19.0 // indirect
	golang.org/x/sys v0.15.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	golang.org/x/tools v0.16.1 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace system-tests => ../system-tests

replace s3-blobstore-backup-restore => ../s3-blobstore-backup-restore
