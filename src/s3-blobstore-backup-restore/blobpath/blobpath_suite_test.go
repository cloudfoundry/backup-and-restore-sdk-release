package blobpath_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestBlobpath(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Blobpath Suite")
}
