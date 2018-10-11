package gcs_test

import (
	"io/ioutil"

	"github.com/onsi/gomega/gexec"

	"fmt"

	. "github.com/onsi/gomega"
)

type GCSClient struct{}

func (c GCSClient) WriteBlobToBucket(bucket, blobName, body string) {
	file, err := ioutil.TempFile("", "bbr-sdk-gcs-system-tests")
	Expect(err).NotTo(HaveOccurred())

	_, err = file.WriteString(body)
	Expect(err).NotTo(HaveOccurred())

	MustRunSuccessfully("gsutil", "cp", file.Name(), fmt.Sprintf("gs://%s/%s", bucket, blobName))
}

func (c GCSClient) ReadBlobFromBucket(bucket, blobName string) string {
	file, err := ioutil.TempFile("", "bbr-sdk-gcs-system-tests")
	Expect(err).NotTo(HaveOccurred())

	MustRunSuccessfully("gsutil", "cp", fmt.Sprintf("gs://%s/%s", bucket, blobName), file.Name())

	body, err := ioutil.ReadFile(file.Name())
	Expect(err).NotTo(HaveOccurred())

	return string(body)
}

func (c GCSClient) DeleteBlobInBucket(bucket, blobName string) {
	MustRunSuccessfully("gsutil", "rm", fmt.Sprintf("gs://%s/%s", bucket, blobName))
}

func (c GCSClient) ListDirsFromBucket(bucket string) string {
	session := Run("gsutil", "ls", fmt.Sprintf("gs://%s/", bucket))
	Eventually(session).Should(gexec.Exit(0))
	return string(session.Out.Contents())
}
