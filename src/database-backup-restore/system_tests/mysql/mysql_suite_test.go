package mysql

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"database/sql"
	"os"
	"testing"

	"github.com/onsi/gomega/gexec"

	"strconv"

	. "database-backup-restore/system_tests/utils"
)

var proxySession *gexec.Session
var connection *sql.DB

var mysqlHostName string
var mysqlNonSslUsername string
var mysqlPassword string
var mysqlPort int

var mysqlCaCert string
var mysqlClientCert string
var mysqlClientKey string

var brJob JobInstance

func TestMysql(t *testing.T) {
	RegisterFailHandler(Fail)
	SetDefaultEventuallyTimeout(15 * time.Minute)
	RunSpecs(t, "Mysql Suite")
}

var _ = Describe("mysql", func() {
	BeforeSuite(func() {
		brJob = JobInstance{
			Deployment:    MustHaveEnv("SDK_DEPLOYMENT"),
			Instance:      MustHaveEnv("SDK_INSTANCE_GROUP"),
			InstanceIndex: "0",
		}

		mysqlHostName = MustHaveEnv("MYSQL_HOSTNAME")
		mysqlNonSslUsername = MustHaveEnv("MYSQL_USERNAME")
		mysqlPassword = MustHaveEnv("MYSQL_PASSWORD")
		mysqlPort, _ = strconv.Atoi(MustHaveEnv("MYSQL_PORT"))

		mysqlCaCert = os.Getenv("MYSQL_CA_CERT")
		mysqlClientCert = os.Getenv("MYSQL_CLIENT_CERT")
		mysqlClientKey = os.Getenv("MYSQL_CLIENT_KEY")

		connection, proxySession = ConnectMysql(
			mysqlHostName,
			mysqlPassword,
			mysqlNonSslUsername,
			mysqlPort,
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

func maybeSkipTLSVerifyIdentityTests() {
	if os.Getenv("TEST_TLS_VERIFY_IDENTITY") == "false" {
		Skip("Skipping TLS verify identity tests")
	}
}
