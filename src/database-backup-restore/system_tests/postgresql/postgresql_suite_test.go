package postgresql

import (
	"os"
	"strconv"
	"testing"
	"time"

	. "database-backup-restore/system_tests/utils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPostgresql(t *testing.T) {
	RegisterFailHandler(Fail)
	SetDefaultEventuallyTimeout(15 * time.Minute)
	RunSpecs(t, "Postgresql Suite")
}

var (
	postgresHostName string
	postgresUsername string
	postgresPassword string
	postgresPort     int

	brJob JobInstance
)

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
})
