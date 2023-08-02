package postgresql

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"

	. "github.com/onsi/gomega"
)

func TestPostgresql(t *testing.T) {
	RegisterFailHandler(Fail)
	SetDefaultEventuallyTimeout(15 * time.Minute)
	RunSpecs(t, "Postgresql Suite")
}
