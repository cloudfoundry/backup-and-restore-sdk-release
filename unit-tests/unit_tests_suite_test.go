package unit_tests

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestUnitTests(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "UnitTests Suite")
}
