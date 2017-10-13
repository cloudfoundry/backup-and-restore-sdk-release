package blobstore_test

import (
	"os"

	. "github.com/cloudfoundry-incubator/blobstore-backup-restore"

	"io/ioutil"

	"path/filepath"

	"encoding/json"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("FileArtifact", func() {
	var backupDir string
	var artifactPath string
	var fileArtifact FileArtifact

	BeforeEach(func() {
		backupDir, _ = ioutil.TempDir("", "bbr_test_")
		artifactPath = filepath.Join(backupDir, "blobstore.json")
		fileArtifact = NewFileArtifact(artifactPath)
	})

	AfterEach(func() {
		os.RemoveAll(backupDir)
	})

	It("Saves the artifact to a file", func() {
		backup := Backup{
			DropletsBackup: BucketBackup{
				BucketName: "my_droplets_bucket",
				RegionName: "my_droplets_region",
				Versions: []LatestVersion{
					{BlobKey: "one", Id: "11"},
					{BlobKey: "two", Id: "21"},
				},
			},
			BuildpacksBackup: BucketBackup{
				BucketName: "my_buildpacks_bucket",
				RegionName: "my_buildpacks_region",
				Versions: []LatestVersion{
					{BlobKey: "three", Id: "31"},
				},
			},
			PackagesBackup: BucketBackup{
				BucketName: "my_packages_bucket",
				RegionName: "my_packages_region",
				Versions: []LatestVersion{
					{BlobKey: "four", Id: "41"},
				},
			},
		}

		err := fileArtifact.Save(backup)
		Expect(err).NotTo(HaveOccurred())

		savedBackup := parseBackupFile(artifactPath)
		Expect(savedBackup).To(Equal(backup))
	})
})

func parseBackupFile(filePath string) Backup {
	fileContents, err := ioutil.ReadFile(filePath)
	Expect(err).NotTo(HaveOccurred())

	var backup Backup
	json.Unmarshal(fileContents, &backup)

	return backup
}
