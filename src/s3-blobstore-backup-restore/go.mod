module s3-blobstore-backup-restore

go 1.20

require (
	github.com/aws/aws-sdk-go v1.44.227
	github.com/cloudfoundry-incubator/bosh-backup-and-restore v1.9.42
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.27.5
	system-tests v0.0.0
)

require (
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/nxadm/tail v1.4.8 // indirect
	golang.org/x/net v0.8.0 // indirect
	golang.org/x/sys v0.6.0 // indirect
	golang.org/x/text v0.8.0 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace system-tests => ../system-tests
