package system_tests

import (
	"io/ioutil"

	"os"
	"strconv"
	"time"

	. "github.com/cloudfoundry-incubator/azure-blobstore-backup-restore/system_tests/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Azure backup and restore", func() {
	var instance JobInstance
	var instanceArtifactDirPath string
	var localArtifact *os.File
	var fileName1, fileName2, fileName3 string
	var containerName string

	BeforeEach(func() {
		instance = JobInstance{
			Deployment:    MustHaveEnv("BOSH_DEPLOYMENT"),
			Instance:      "azure-backuper",
			InstanceIndex: "0",
		}

		containerName = ContainerName()
		CreateContainer(containerName)

		fileName1 = "test_file_1_" + strconv.FormatInt(time.Now().Unix(), 10)
		fileName2 = "test_file_2_" + strconv.FormatInt(time.Now().Unix(), 10)
		fileName3 = "test_file_3_" + strconv.FormatInt(time.Now().Unix(), 10)

		instanceArtifactDirPath = "/tmp/azure-blobstore-backup-restorer" + strconv.FormatInt(time.Now().Unix(), 10)
		instance.RunOnInstanceAndSucceed("mkdir -p " + instanceArtifactDirPath)
		var err error
		localArtifact, err = ioutil.TempFile("", "blobstore-")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		instance.RunOnInstanceAndSucceed("sudo rm -rf " + instanceArtifactDirPath)
		err := os.Remove(localArtifact.Name())
		Expect(err).NotTo(HaveOccurred())

		DeleteContainer(containerName)
	})

	It("backs up successfully", func() {
		WriteFileInContainer(containerName, fileName1, "TEST_BLOB_1_OLD")
		WriteFileInContainer(containerName, fileName1, "TEST_BLOB_1")
		WriteFileInContainer(containerName, fileName2, "TEST_BLOB_2_OLDEST")
		WriteFileInContainer(containerName, fileName2, "TEST_BLOB_2_OLD")
		WriteFileInContainer(containerName, fileName2, "TEST_BLOB_2")
		WriteFileInContainer(containerName, fileName3, "TEST_BLOB_3")

		instance.RunOnInstanceAndSucceed("BBR_ARTIFACT_DIRECTORY=" + instanceArtifactDirPath + " /var/vcap/jobs/azure-blobstore-backup-restorer/bin/bbr/backup")
		instance.DownloadFromInstanceAndSucceed(instanceArtifactDirPath+"/blobstore.json", localArtifact.Name())

		fileContents, err := ioutil.ReadFile(localArtifact.Name())

		Expect(err).NotTo(HaveOccurred())
		Expect(fileContents).To(ContainSubstring("\"name\":\"" + containerName + "\""))
		Expect(fileContents).To(ContainSubstring("\"name\":\"" + fileName1 + "\",\"hash\":\"R1M39xrrgP7eS+jJHBWu1A==\""))
		Expect(fileContents).To(ContainSubstring("\"name\":\"" + fileName2 + "\",\"hash\":\"L+IcKub+0Og4CXjKqA1/3w==\""))
		Expect(fileContents).To(ContainSubstring("\"name\":\"" + fileName3 + "\",\"hash\":\"7VBVkm19ll+P6THGtqGHww==\""))
	})
})
