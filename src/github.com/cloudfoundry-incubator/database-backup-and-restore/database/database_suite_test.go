package database

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestDatabase(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Database Suite")
}
