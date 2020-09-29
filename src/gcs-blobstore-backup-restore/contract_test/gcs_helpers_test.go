package contract_test

import (
	"os"

	"os/exec"

	"time"

	"fmt"

	"io/ioutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func MustHaveEnv(keyname string) string {
	val := os.Getenv(keyname)
	Expect(val).NotTo(BeEmpty(), "Need "+keyname+" for the test")
	return val
}

func Authenticate(serviceAccountKey, projectName string) {
	tmpFile, err := ioutil.TempFile("", "gcp_service_account_key_")
	Expect(err).NotTo(HaveOccurred())
	err = ioutil.WriteFile(tmpFile.Name(), []byte(serviceAccountKey), 0644)
	Expect(err).NotTo(HaveOccurred())
	runSuccessfully("gcloud", "auth", "activate-service-account", "--key-file", tmpFile.Name(), "--project", projectName)
}

func CreateBucketWithTimestampedName(prefix string) string {
	bucketName := fmt.Sprintf("contract-test-%s-%d", prefix, time.Now().UnixNano())
	runSuccessfully("gsutil", "mb", "-c", "regional", "-l", "europe-west2", "gs://"+bucketName)
	return bucketName
}

func DeleteBucket(bucketName string) {
	runSuccessfully("gsutil", "rm", "-r", "gs://"+bucketName)
}

func UploadFile(bucketName, blobName, fileContents string) {
	file := createTmpFile("", blobName, fileContents)
	runSuccessfully("gsutil", "cp", "-v", file.Name(), "gs://"+bucketName+"/"+blobName)
}

func UploadFileWithDir(bucketName, dir, blobName, fileContents string) {
	file := createTmpFile("", blobName, fileContents)
	runSuccessfully("gsutil", "cp", "-v", file.Name(), "gs://"+bucketName+"/"+dir+"/"+blobName)
}

func createTmpFile(dirName, fileName, fileContents string) *os.File {
	dir, err := ioutil.TempDir("", dirName)
	Expect(err).NotTo(HaveOccurred())
	file, err := ioutil.TempFile(dir, fileName)
	Expect(err).NotTo(HaveOccurred())
	err = ioutil.WriteFile(file.Name(), []byte(fileContents), 0644)
	Expect(err).NotTo(HaveOccurred())
	return file
}

func runSuccessfully(command string, args ...string) (string, string) {
	cmd := exec.Command(command, args...)
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gexec.Exit(0))
	return string(session.Out.Contents()), string(session.Err.Contents())
}
