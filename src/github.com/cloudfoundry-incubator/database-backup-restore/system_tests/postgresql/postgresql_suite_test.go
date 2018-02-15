package postgresql

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"os"
	"testing"

	"strconv"

	. "github.com/cloudfoundry-incubator/database-backup-restore/system_tests/utils"
)

var pgConnection *PostgresConnection

var postgresHostName string
var postgresNonSslUsername string
var postgresSslUsername string
var postgresMutualTlsUsername string
var postgresPassword string
var postgresPort int

var postgresCaCert string
var postgresClientCert string
var postgresClientKey string

var brJob JobInstance

func TestPostgresql(t *testing.T) {
	RegisterFailHandler(Fail)
	SetDefaultEventuallyTimeout(15 * time.Minute)
	RunSpecs(t, "Postgresql Suite")
}

var _ = Describe("postgres", func() {
	BeforeSuite(func() {
		brJob = JobInstance{
			Deployment:    MustHaveEnv("SDK_DEPLOYMENT"),
			Instance:      MustHaveEnv("SDK_INSTANCE_GROUP"),
			InstanceIndex: "0",
		}

		postgresHostName = MustHaveEnv("POSTGRES_HOSTNAME")
		postgresPassword = MustHaveEnv("POSTGRES_PASSWORD")
		postgresNonSslUsername = MustHaveEnv("POSTGRES_USERNAME")
		postgresSslUsername = os.Getenv("POSTGRES_SSL_USERNAME")
		postgresMutualTlsUsername = os.Getenv("POSTGRES_MUTUAL_TLS_USERNAME")
		postgresPort, _ = strconv.Atoi(MustHaveEnv("POSTGRES_PORT"))

		postgresCaCert = os.Getenv("POSTGRES_CA_CERT")
		postgresClientCert = os.Getenv("POSTGRES_CLIENT_CERT")
		postgresClientKey = os.Getenv("POSTGRES_CLIENT_KEY")
	})
})
