package system_tests

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
	"time"
)

func TestSystemTests(t *testing.T) {
	RegisterFailHandler(Fail)
	SetDefaultEventuallyTimeout(15 * time.Minute)
	RunSpecs(t, "SystemTests Suite")
}
