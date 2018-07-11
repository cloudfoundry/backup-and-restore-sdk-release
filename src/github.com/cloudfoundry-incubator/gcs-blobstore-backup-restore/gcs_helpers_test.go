package gcs_test

import (
	"os"

	"os/exec"

	"time"

	"fmt"

	"io/ioutil"

	"strings"

	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

func MustHaveEnv(keyname string) string {
	val := os.Getenv(keyname)
	Expect(val).NotTo(BeEmpty(), "Need "+keyname+" for the test")
	return val
}

func Authenticate(serviceAccountKeyPath string) {
	runSuccessfully("gcloud", "auth", "activate-service-account", "--key-file", serviceAccountKeyPath)
}

func CreateBucketWithTimestampedName(prefix string, versioned bool) string {
	bucketName := fmt.Sprintf("%s_%d", prefix, time.Now().UnixNano())
	runSuccessfully("gsutil", "mb", "-c", "regional", "-l", "europe-west2", "gs://"+bucketName)
	setVersioning(bucketName, versioned)
	return bucketName
}

func DeleteBucket(bucketName string) {
	runSuccessfully("gsutil", "rm", "-r", "gs://"+bucketName)
}

func setVersioning(bucketName string, versioned bool) {
	var versioning string
	if versioned {
		versioning = "on"
	} else {
		versioning = "off"
	}
	runSuccessfully("gsutil", "versioning", "set", versioning, "gs://"+bucketName)
}

func UploadFile(bucketName, fileName, fileContents string) int64 {
	file := createTmpFile(fileName, fileContents)

	_, stdErr := runSuccessfully("gsutil", "cp", "-v", file.Name(), "gs://"+bucketName+"/"+fileName)
	generationIdString := strings.TrimSpace(strings.Split(strings.Split(stdErr, "\n")[1], "#")[1])
	generationId, err := strconv.ParseInt(generationIdString, 10, 64)
	Expect(err).NotTo(HaveOccurred())

	return generationId
}

func createTmpFile(fileName, fileContents string) *os.File {
	file, err := ioutil.TempFile("", fileName)
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
