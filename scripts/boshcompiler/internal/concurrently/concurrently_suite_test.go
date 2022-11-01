package concurrently_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestConcurrently(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Concurrently Suite")
}
