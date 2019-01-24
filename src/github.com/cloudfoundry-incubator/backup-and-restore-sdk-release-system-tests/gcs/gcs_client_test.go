package gcs_test

import (
	"io/ioutil"
	"strconv"
	"time"

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

func (c GCSClient) WriteNBlobsToBucket(bucket string, blobName string, blobBody string, n int) {
	blobsDir, err := ioutil.TempDir("", "testdir")
	Expect(err).NotTo(HaveOccurred())
	for i := 0; i < n; i++ {
		timestampedName := blobName + strconv.FormatInt(time.Now().Unix(), 10)
		file, err := ioutil.TempFile(blobsDir, fmt.Sprintf(timestampedName, i))
		Expect(err).NotTo(HaveOccurred())
		_, err = file.WriteString(fmt.Sprintf(blobBody, i))
		Expect(err).NotTo(HaveOccurred())
	}

	MustRunSuccessfully("gsutil", "-m", "-q", "cp", "-r", blobsDir+"/*", fmt.Sprintf("gs://%s", bucket))
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

func (c GCSClient) DeleteAllBlobInBucket(bucket string) {
	session := Run("gsutil", "-m", "-q", "rm", "-r", fmt.Sprintf("gs://%s", bucket))
	Eventually(session).Should(gexec.Exit())
}

func (c GCSClient) ListDirsFromBucket(bucket string) string {
	session := Run("gsutil", "ls", fmt.Sprintf("gs://%s/", bucket))
	Eventually(session).Should(gexec.Exit(0))
	return string(session.Out.Contents())
}

func (c GCSClient) ListBlobsFromBucket(bucket string) string {
	session := Run("gsutil", "ls", fmt.Sprintf("gs://%s/**/*", bucket))
	Eventually(session).Should(gexec.Exit(0))
	return string(session.Out.Contents())
}

func (c GCSClient) WriteNSizeBlobToBucket(bucket string, blobName string, size int) {
	blobsDir, err := ioutil.TempDir("", "testdir")
	Expect(err).NotTo(HaveOccurred())
	fileToCopy := fmt.Sprintf("%s/%s", blobsDir, blobName)
	gigabyte := 1024 * 1024 * 1024
	MustRunSuccessfully(
		"dd",
		"if=/dev/zero",
		fmt.Sprintf("of=%s", fileToCopy),
		fmt.Sprintf("bs=%d", gigabyte),
		fmt.Sprintf("count=%d", size),
	)

	MustRunSuccessfully("gsutil", "-o", "GSUtil:parallel_composite_upload_threshold=150M", "cp", fileToCopy, fmt.Sprintf("gs://%s/%s", bucket, blobName))

}
