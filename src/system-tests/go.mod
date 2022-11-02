module system-tests

go 1.18

require (
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.19.0
	s3-blobstore-backup-restore v0.0.0
)

require (
	github.com/cloudfoundry-incubator/bosh-backup-and-restore v1.9.27 // indirect
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/nxadm/tail v1.4.8 // indirect
	golang.org/x/net v0.0.0-20220225172249-27dd8689420f // indirect
	golang.org/x/sys v0.0.0-20220310020820-b874c991c1a5 // indirect
	golang.org/x/text v0.3.7 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace s3-blobstore-backup-restore => ../s3-blobstore-backup-restore
