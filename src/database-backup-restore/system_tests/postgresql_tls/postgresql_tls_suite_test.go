package postgresql_tls_test

import (
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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
