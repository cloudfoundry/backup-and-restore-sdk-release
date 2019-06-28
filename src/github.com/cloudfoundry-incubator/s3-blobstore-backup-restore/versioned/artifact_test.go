package versioned_test

import (
	"os"

	"io/ioutil"

	"path/filepath"

	"github.com/cloudfoundry-incubator/backup-and-restore-sdk/s3-blobstore-backup-restore/versioned"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("FileArtifact", func() {
	var backupDir string
	var artifactPath string
	var fileArtifact versioned.FileArtifact

	BeforeEach(func() {
		backupDir, _ = ioutil.TempDir("", "bbr_test_")
		artifactPath = filepath.Join(backupDir, "blobstore.json")
		fileArtifact = versioned.NewFileArtifact(artifactPath)
	})

	AfterEach(func() {
		os.RemoveAll(backupDir)
	})

	It("saves the artifact to a file", func() {
		backup := map[string]versioned.BucketSnapshot{
			"droplets": {
				BucketName: "my_droplets_bucket",
				RegionName: "my_droplets_region",
				Versions: []versioned.BlobVersion{
					{BlobKey: "one", Id: "11"},
					{BlobKey: "two", Id: "21"},
				},
			},
			"buildpacks": {
				BucketName: "my_buildpacks_bucket",
				RegionName: "my_buildpacks_region",
				Versions: []versioned.BlobVersion{
					{BlobKey: "three", Id: "31"},
				},
			},
			"packages": {
				BucketName: "my_packages_bucket",
				RegionName: "my_packages_region",
				Versions: []versioned.BlobVersion{
					{BlobKey: "four", Id: "41"},
				},
			},
		}

		err := fileArtifact.Save(backup)
		Expect(err).NotTo(HaveOccurred())

		savedBackup, err := fileArtifact.Load()
		Expect(err).NotTo(HaveOccurred())
		Expect(savedBackup).To(Equal(backup))
	})

	Context("when saving the file fails", func() {
		BeforeEach(func() {
			fileArtifact = versioned.NewFileArtifact("/this/path/does/not/exist")
		})

		It("returns an error", func() {
			err := fileArtifact.Save(map[string]versioned.BucketSnapshot{})
			Expect(err).To(MatchError(ContainSubstring("could not write backup file")))
		})
	})

	Context("when reading the file fails", func() {
		BeforeEach(func() {
			fileArtifact = versioned.NewFileArtifact("/this/path/does/not/exist")
		})

		It("returns an error", func() {
			backup, err := fileArtifact.Load()
			Expect(backup).To(BeNil())
			Expect(err).To(MatchError(ContainSubstring("could not read backup file")))
		})
	})

	Context("when the artifact has an invalid format", func() {
		BeforeEach(func() {
			ioutil.WriteFile(artifactPath, []byte("THIS IS NOT VALID JSON"), 0666)
		})

		It("returns an error", func() {
			backup, err := fileArtifact.Load()
			Expect(backup).To(BeNil())
			Expect(err).To(MatchError(ContainSubstring("backup file has an invalid format")))
		})
	})
})
