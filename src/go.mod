module github.com/cloudfoundry/backup-and-restore-sdk-release/src

go 1.16

require (
	github.com/aws/aws-sdk-go v1.43.20
	github.com/cloudfoundry-incubator/bosh-backup-and-restore v1.9.25
	github.com/onsi/ginkgo v1.16.5

	github.com/Azure/azure-storage-blob-go v0.14.0
	github.com/onsi/gomega v1.19.0
	github.com/go-sql-driver/mysql v1.6.0
	github.com/lib/pq v1.10.4
	github.com/pivotal-cf-experimental/go-binmock v0.0.0-20171027112700-f797157c64e9
	github.com/pivotal-cf/go-binmock v0.0.0-20171027112700-f797157c64e9 // indirect

	cloud.google.com/go/storage v1.21.0
	github.com/cloudfoundry-incubator/bosh-backup-and-restore v1.9.25
	golang.org/x/net v0.0.0-20220225172249-27dd8689420f
	golang.org/x/oauth2 v0.0.0-20220309155454-6242fa91716a
	google.golang.org/api v0.73.0
)

