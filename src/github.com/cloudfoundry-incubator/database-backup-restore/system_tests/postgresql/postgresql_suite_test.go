package postgresql

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"os"
	"testing"

	"database/sql"

	. "github.com/cloudfoundry-incubator/database-backup-restore/system_tests/utils"
	"github.com/onsi/gomega/gexec"
)

var proxySession *gexec.Session
var connection *sql.DB

var postgresHostName string
var postgresNonSslUsername string
var postgresSslUsername string
var postgresPassword string
var postgresPort string

var postgresCaCert string

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
		postgresPort = MustHaveEnv("POSTGRES_PORT")

		postgresCaCert = os.Getenv("POSTGRES_CA_CERT")

		connection, proxySession = SuccessfullyConnectToPostgres(
			postgresHostName,
			postgresPassword,
			postgresNonSslUsername,
			postgresPort,
			"postgres",
			os.Getenv("SSH_PROXY_HOST"),
			os.Getenv("SSH_PROXY_USER"),
			os.Getenv("SSH_PROXY_KEY_FILE"),
		)
	})

	AfterSuite(func() {
		if proxySession != nil {
			proxySession.Kill()
		}
	})
})
