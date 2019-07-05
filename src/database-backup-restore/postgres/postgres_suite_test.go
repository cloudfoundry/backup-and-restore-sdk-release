package postgres_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestPostgres(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Postgres Suite")
}
