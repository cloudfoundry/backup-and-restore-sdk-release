package unversioned_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestUnversioned(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Unversioned Suite")
}
