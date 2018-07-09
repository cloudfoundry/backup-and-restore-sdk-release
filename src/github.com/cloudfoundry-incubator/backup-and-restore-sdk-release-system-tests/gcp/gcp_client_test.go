package gcp_test

import (
	"io/ioutil"

	"fmt"

	. "github.com/onsi/gomega"
)

type GCPClient struct{}

func (c GCPClient) WriteBlobToBucket(bucket, blobName, body string) {
	file, err := ioutil.TempFile("", "bbr-sdk-gcp-system-tests")
	Expect(err).NotTo(HaveOccurred())

	_, err = file.WriteString(body)
	Expect(err).NotTo(HaveOccurred())

	MustRunSuccessfully("gsutil", "cp", file.Name(), fmt.Sprintf("gs://%s/%s", bucket, blobName))
}

func (c GCPClient) DeleteBlobInBucket(bucket, blobName string) {
	MustRunSuccessfully("gsutil", "rm", fmt.Sprintf("gs://%s/%s", bucket, blobName))
}
