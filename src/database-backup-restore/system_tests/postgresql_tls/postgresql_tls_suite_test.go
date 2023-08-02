package postgresql_tls_test

import (
	"os"
	"strconv"
	"testing"
	"time"

	. "database-backup-restore/system_tests/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	postgresHostName string
	postgresUsername string
	postgresPassword string
	postgresPort     int
	postgresCaCert   string

	brJob JobInstance
)

func TestPostgresqlTls(t *testing.T) {
	RegisterFailHandler(Fail)
	SetDefaultEventuallyTimeout(15 * time.Minute)
	RunSpecs(t, "PostgresqlTls Suite")
}

func maybeSkipTLSVerifyIdentityTests() {
	if os.Getenv("TEST_TLS_VERIFY_IDENTITY") == "false" {
		Skip("Skipping TLS verify identity tests")
	}
}

var _ = BeforeSuite(func() {
	if os.Getenv("RUN_TESTS_WITHOUT_BOSH") != "true" {
		brJob = JobInstance{
			Deployment:    MustHaveEnv("SDK_DEPLOYMENT"),
			Instance:      MustHaveEnv("SDK_INSTANCE_GROUP"),
			InstanceIndex: "0",
		}
	}

	postgresHostName = MustHaveEnv("POSTGRES_HOSTNAME")
	postgresPort, _ = strconv.Atoi(MustHaveEnv("POSTGRES_PORT"))
	postgresPassword = MustHaveEnv("POSTGRES_PASSWORD")
	postgresUsername = MustHaveEnv("POSTGRES_USERNAME")
	postgresCaCert = os.Getenv("POSTGRES_CA_CERT")
})
