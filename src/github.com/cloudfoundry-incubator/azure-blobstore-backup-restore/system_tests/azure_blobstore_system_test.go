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
	var localArtifactDirectory string
	var fileName1, fileName2, fileName3 string
	var containerName string

	BeforeEach(func() {
		instance = JobInstance{
			Deployment:    MustHaveEnv("BOSH_DEPLOYMENT"),
			Instance:      "azure-backuper",
			InstanceIndex: "0",
		}

		containerName = MustHaveEnv("AZURE_CONTAINER_NAME")

		fileName1 = "test_file_1_" + strconv.FormatInt(time.Now().Unix(), 10)
		fileName2 = "test_file_2_" + strconv.FormatInt(time.Now().Unix(), 10)
		fileName3 = "test_file_3_" + strconv.FormatInt(time.Now().Unix(), 10)

		instanceArtifactDirPath = "/tmp/azure-blobstore-backup-restorer" + strconv.FormatInt(time.Now().Unix(), 10)
		instance.RunOnInstanceAndSucceed("mkdir -p " + instanceArtifactDirPath)
		var err error
		localArtifactDirectory, err = ioutil.TempDir("", "azure-blobstore-")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		instance.RunOnInstanceAndSucceed("sudo rm -rf " + instanceArtifactDirPath)
		err := os.RemoveAll(localArtifactDirectory)
		Expect(err).NotTo(HaveOccurred())

		DeleteFileInContainer(containerName, fileName1)
		DeleteFileInContainer(containerName, fileName2)
		DeleteFileInContainer(containerName, fileName3)
	})

	It("backs up and restores in-place successfully", func() {
		WriteFileInContainer(containerName, fileName1, "TEST_BLOB_1")
		WriteFileInContainer(containerName, fileName2, "TEST_BLOB_2")
		WriteFileInContainer(containerName, fileName3, "TEST_BLOB_3")

		instance.RunOnInstanceAndSucceed("BBR_ARTIFACT_DIRECTORY=" + instanceArtifactDirPath + " /var/vcap/jobs/azure-blobstore-backup-restorer/bin/bbr/backup")

		WriteFileInContainer(containerName, fileName2, "TEST_BLOB_2_NEW")
		DeleteFileInContainer(containerName, fileName3)

		instance.RunOnInstanceAndSucceed("BBR_ARTIFACT_DIRECTORY=" + instanceArtifactDirPath + " /var/vcap/jobs/azure-blobstore-backup-restorer/bin/bbr/restore")

		Expect(ReadFileFromContainer(containerName, fileName1)).To(Equal("TEST_BLOB_1"))
		Expect(ReadFileFromContainer(containerName, fileName2)).To(Equal("TEST_BLOB_2"))
		Expect(ReadFileFromContainer(containerName, fileName3)).To(Equal("TEST_BLOB_3"))
	})

	Context("when the destination container is different from the source container", func() {
		var restoreInstance JobInstance
		var cloneContainerName string

		BeforeEach(func() {
			restoreInstance = JobInstance{
				Deployment:    MustHaveEnv("BOSH_DEPLOYMENT"),
				Instance:      "azure-restore-to-clone",
				InstanceIndex: "0",
			}
			cloneContainerName = MustHaveEnv("AZURE_CLONE_CONTAINER_NAME")
			restoreInstance.RunOnInstanceAndSucceed("mkdir -p " + instanceArtifactDirPath)
		})

		AfterEach(func() {
			restoreInstance.RunOnInstanceAndSucceed("sudo rm -rf " + instanceArtifactDirPath)
			err := os.RemoveAll(localArtifactDirectory)
			Expect(err).NotTo(HaveOccurred())

			DeleteFileInContainer(cloneContainerName, fileName1)
			DeleteFileInContainer(cloneContainerName, fileName2)
			DeleteFileInContainer(cloneContainerName, fileName3)
		})

		It("backs up and restores cloned container successfully", func() {
			WriteFileInContainer(containerName, fileName1, "TEST_BLOB_1")
			WriteFileInContainer(containerName, fileName2, "TEST_BLOB_2")
			WriteFileInContainer(containerName, fileName3, "TEST_BLOB_3")

			instance.RunOnInstanceAndSucceed("BBR_ARTIFACT_DIRECTORY=" + instanceArtifactDirPath + " /var/vcap/jobs/azure-blobstore-backup-restorer/bin/bbr/backup")

			WriteFileInContainer(containerName, fileName2, "TEST_BLOB_2_NEW")
			DeleteFileInContainer(containerName, fileName3)

			instance.DownloadFromInstance(instanceArtifactDirPath+"/blobstore.json", localArtifactDirectory)
			restoreInstance.UploadToInstance(localArtifactDirectory+"/blobstore.json", instanceArtifactDirPath)

			restoreInstance.RunOnInstanceAndSucceed("BBR_ARTIFACT_DIRECTORY=" + instanceArtifactDirPath + " /var/vcap/jobs/azure-blobstore-backup-restorer/bin/bbr/restore")

			Expect(ReadFileFromContainer(cloneContainerName, fileName1)).To(Equal("TEST_BLOB_1"))
			Expect(ReadFileFromContainer(cloneContainerName, fileName2)).To(Equal("TEST_BLOB_2"))
			Expect(ReadFileFromContainer(cloneContainerName, fileName3)).To(Equal("TEST_BLOB_3"))
		})
	})
})
