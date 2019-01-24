package incremental_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestIncremental(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Incremental Suite")
}
