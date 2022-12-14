package azure_test

import (
	"os"
	"strconv"
	"time"

	. "system-tests"
	. "system-tests/azure"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const endpointSuffix = "core.windows.net"

var _ = Describe("Azure backup and restore", func() {
	var azureClient AzureClient
	var instance JobInstance
	var instanceArtifactDirPath string
	var localArtifactDirectory string
	var fileName1, fileName2, fileName3 string
	var containerName string

	BeforeEach(func() {
		azureClient = NewAzureClient(
			MustHaveEnv("AZURE_STORAGE_ACCOUNT"),
			MustHaveEnv("AZURE_STORAGE_KEY"),
			endpointSuffix,
		)

		containerName = MustHaveEnv("AZURE_CONTAINER_NAME")

		fileName1 = "test_file_1_" + strconv.FormatInt(time.Now().Unix(), 10)
		fileName2 = "test_file_2_" + strconv.FormatInt(time.Now().Unix(), 10)
		fileName3 = "test_file_3_" + strconv.FormatInt(time.Now().Unix(), 10)

		var err error
		localArtifactDirectory, err = os.MkdirTemp("", "azure-blobstore-")
		Expect(err).NotTo(HaveOccurred())

		instanceArtifactDirPath = "/tmp/azure-blobstore-backup-restorer" + strconv.FormatInt(time.Now().Unix(), 10)
	})

	Context("When BPM is enabled", func() {
		BeforeEach(func() {
			instance = JobInstance{
				Deployment: MustHaveEnv("BOSH_DEPLOYMENT"),
				Name:       "azure-backuper-bpm",
				Index:      "0",
			}

			instanceArtifactDirPath = "/var/vcap/store/azure-blobstore-backup-restorer" + strconv.FormatInt(time.Now().Unix(), 10)
			instance.RunSuccessfully("sudo mkdir -p " + instanceArtifactDirPath)

		})

		AfterEach(func() {
			instance.RunSuccessfully("sudo rm -rf " + instanceArtifactDirPath)
			err := os.RemoveAll(localArtifactDirectory)
			Expect(err).NotTo(HaveOccurred())

			azureClient.DeleteFileInContainer(containerName, fileName1)
			azureClient.DeleteFileInContainer(containerName, fileName2)
			azureClient.DeleteFileInContainer(containerName, fileName3)
		})

		It("backs up and restores successfully", func() {
			azureClient.WriteFileInContainer(containerName, fileName1, "TEST_BLOB_1")
			azureClient.WriteFileInContainer(containerName, fileName2, "TEST_BLOB_2")
			azureClient.WriteFileInContainer(containerName, fileName3, "TEST_BLOB_3")

			instance.RunSuccessfully("sudo BBR_ARTIFACT_DIRECTORY=" + instanceArtifactDirPath + " /var/vcap/jobs/azure-blobstore-backup-restorer/bin/bbr/backup")

			azureClient.WriteFileInContainer(containerName, fileName2, "TEST_BLOB_2_NEW")
			azureClient.DeleteFileInContainer(containerName, fileName3)

			instance.RunSuccessfully("sudo BBR_ARTIFACT_DIRECTORY=" + instanceArtifactDirPath + " /var/vcap/jobs/azure-blobstore-backup-restorer/bin/bbr/restore")

			Expect(azureClient.ReadFileFromContainer(containerName, fileName1)).To(Equal("TEST_BLOB_1"))
			Expect(azureClient.ReadFileFromContainer(containerName, fileName2)).To(Equal("TEST_BLOB_2"))
			Expect(azureClient.ReadFileFromContainer(containerName, fileName3)).To(Equal("TEST_BLOB_3"))
		})
	})
	Context("when the destination container is the same as the source container", func() {
		BeforeEach(func() {
			instance = JobInstance{
				Deployment: MustHaveEnv("BOSH_DEPLOYMENT"),
				Name:       "azure-backuper",
				Index:      "0",
			}

			instance.RunSuccessfully("mkdir -p " + instanceArtifactDirPath)
		})

		AfterEach(func() {
			instance.RunSuccessfully("sudo rm -rf " + instanceArtifactDirPath)
			err := os.RemoveAll(localArtifactDirectory)
			Expect(err).NotTo(HaveOccurred())

			azureClient.DeleteFileInContainer(containerName, fileName1)
			azureClient.DeleteFileInContainer(containerName, fileName2)
			azureClient.DeleteFileInContainer(containerName, fileName3)
		})

		It("backs up and restores in-place successfully", func() {
			azureClient.WriteFileInContainer(containerName, fileName1, "TEST_BLOB_1")
			azureClient.WriteFileInContainer(containerName, fileName2, "TEST_BLOB_2")
			azureClient.WriteFileInContainer(containerName, fileName3, "TEST_BLOB_3")

			instance.RunSuccessfully("BBR_ARTIFACT_DIRECTORY=" + instanceArtifactDirPath + " /var/vcap/jobs/azure-blobstore-backup-restorer/bin/bbr/backup")

			azureClient.WriteFileInContainer(containerName, fileName2, "TEST_BLOB_2_NEW")
			azureClient.DeleteFileInContainer(containerName, fileName3)

			instance.RunSuccessfully("BBR_ARTIFACT_DIRECTORY=" + instanceArtifactDirPath + " /var/vcap/jobs/azure-blobstore-backup-restorer/bin/bbr/restore")

			Expect(azureClient.ReadFileFromContainer(containerName, fileName1)).To(Equal("TEST_BLOB_1"))
			Expect(azureClient.ReadFileFromContainer(containerName, fileName2)).To(Equal("TEST_BLOB_2"))
			Expect(azureClient.ReadFileFromContainer(containerName, fileName3)).To(Equal("TEST_BLOB_3"))
		})
	})

	Context("when the destination container is different from the source container", func() {
		var restoreInstance JobInstance
		var differentContainerName string

		BeforeEach(func() {
			instance = JobInstance{
				Deployment: MustHaveEnv("BOSH_DEPLOYMENT"),
				Name:       "azure-backuper",
				Index:      "0",
			}

			instance.RunSuccessfully("mkdir -p " + instanceArtifactDirPath)

			restoreInstance = JobInstance{
				Deployment: MustHaveEnv("BOSH_DEPLOYMENT"),
				Name:       "azure-restore-to-different-container",
				Index:      "0",
			}
			differentContainerName = MustHaveEnv("AZURE_DIFFERENT_CONTAINER_NAME")
			restoreInstance.RunSuccessfully("mkdir -p " + instanceArtifactDirPath)
		})

		AfterEach(func() {
			restoreInstance.RunSuccessfully("sudo rm -rf " + instanceArtifactDirPath)
			err := os.RemoveAll(localArtifactDirectory)
			Expect(err).NotTo(HaveOccurred())

			azureClient.DeleteFileInContainer(containerName, fileName1)
			azureClient.DeleteFileInContainer(containerName, fileName2)
			azureClient.DeleteFileInContainer(containerName, fileName3)
			azureClient.DeleteFileInContainer(differentContainerName, fileName1)
			azureClient.DeleteFileInContainer(differentContainerName, fileName2)
			azureClient.DeleteFileInContainer(differentContainerName, fileName3)
		})

		It("backs up and restores cloned container successfully", func() {
			azureClient.WriteFileInContainer(containerName, fileName1, "TEST_BLOB_1")
			azureClient.WriteFileInContainer(containerName, fileName2, "TEST_BLOB_2")
			azureClient.WriteFileInContainer(containerName, fileName3, "TEST_BLOB_3")

			instance.RunSuccessfully("BBR_ARTIFACT_DIRECTORY=" + instanceArtifactDirPath + " /var/vcap/jobs/azure-blobstore-backup-restorer/bin/bbr/backup")

			azureClient.WriteFileInContainer(containerName, fileName2, "TEST_BLOB_2_NEW")
			azureClient.DeleteFileInContainer(containerName, fileName3)

			instance.Download(instanceArtifactDirPath+"/blobstore.json", localArtifactDirectory)
			restoreInstance.Upload(localArtifactDirectory+"/blobstore.json", instanceArtifactDirPath)

			restoreInstance.RunSuccessfully("BBR_ARTIFACT_DIRECTORY=" + instanceArtifactDirPath + " /var/vcap/jobs/azure-blobstore-backup-restorer/bin/bbr/restore")

			Expect(azureClient.ReadFileFromContainer(differentContainerName, fileName1)).To(Equal("TEST_BLOB_1"))
			Expect(azureClient.ReadFileFromContainer(differentContainerName, fileName2)).To(Equal("TEST_BLOB_2"))
			Expect(azureClient.ReadFileFromContainer(differentContainerName, fileName3)).To(Equal("TEST_BLOB_3"))
		})
	})

	Context("when the destination storage account is different from the source storage account", func() {
		var differentAzureClient AzureClient
		var restoreInstance JobInstance
		var differentContainerName string

		BeforeEach(func() {
			instance = JobInstance{
				Deployment: MustHaveEnv("BOSH_DEPLOYMENT"),
				Name:       "azure-backuper",
				Index:      "0",
			}

			instance.RunSuccessfully("mkdir -p " + instanceArtifactDirPath)

			differentAzureClient = NewAzureClient(
				MustHaveEnv("AZURE_DIFFERENT_STORAGE_ACCOUNT"),
				MustHaveEnv("AZURE_DIFFERENT_STORAGE_KEY"),
				endpointSuffix,
			)

			restoreInstance = JobInstance{
				Deployment: MustHaveEnv("BOSH_DEPLOYMENT"),
				Name:       "azure-restore-to-different-storage-account",
				Index:      "0",
			}
			differentContainerName = MustHaveEnv("AZURE_DIFFERENT_CONTAINER_NAME")
			restoreInstance.RunSuccessfully("mkdir -p " + instanceArtifactDirPath)
		})

		AfterEach(func() {
			restoreInstance.RunSuccessfully("sudo rm -rf " + instanceArtifactDirPath)
			err := os.RemoveAll(localArtifactDirectory)
			Expect(err).NotTo(HaveOccurred())

			azureClient.DeleteFileInContainer(containerName, fileName1)
			azureClient.DeleteFileInContainer(containerName, fileName2)
			azureClient.DeleteFileInContainer(containerName, fileName3)
			differentAzureClient.DeleteFileInContainer(differentContainerName, fileName1)
			differentAzureClient.DeleteFileInContainer(differentContainerName, fileName2)
			differentAzureClient.DeleteFileInContainer(differentContainerName, fileName3)
		})

		It("backs up and restores cloned container successfully", func() {
			azureClient.WriteFileInContainer(containerName, fileName1, "TEST_BLOB_1")
			azureClient.WriteFileInContainer(containerName, fileName2, "TEST_BLOB_2")
			azureClient.WriteFileInContainer(containerName, fileName3, "TEST_BLOB_3")

			instance.RunSuccessfully("BBR_ARTIFACT_DIRECTORY=" + instanceArtifactDirPath + " /var/vcap/jobs/azure-blobstore-backup-restorer/bin/bbr/backup")

			azureClient.WriteFileInContainer(containerName, fileName2, "TEST_BLOB_2_NEW")
			azureClient.DeleteFileInContainer(containerName, fileName3)

			instance.Download(instanceArtifactDirPath+"/blobstore.json", localArtifactDirectory)
			restoreInstance.Upload(localArtifactDirectory+"/blobstore.json", instanceArtifactDirPath)

			restoreInstance.RunSuccessfully("BBR_ARTIFACT_DIRECTORY=" + instanceArtifactDirPath + " /var/vcap/jobs/azure-blobstore-backup-restorer/bin/bbr/restore")

			Expect(differentAzureClient.ReadFileFromContainer(differentContainerName, fileName1)).To(Equal("TEST_BLOB_1"))
			Expect(differentAzureClient.ReadFileFromContainer(differentContainerName, fileName2)).To(Equal("TEST_BLOB_2"))
			Expect(differentAzureClient.ReadFileFromContainer(differentContainerName, fileName3)).To(Equal("TEST_BLOB_3"))
		})
	})
})
