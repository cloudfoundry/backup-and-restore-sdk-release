package gcp_test

import (
	"testing"

	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestGcp(t *testing.T) {
	RegisterFailHandler(Fail)
	SetDefaultEventuallyTimeout(15 * time.Minute)
	RunSpecs(t, "GCP System Tests Suite")
}
