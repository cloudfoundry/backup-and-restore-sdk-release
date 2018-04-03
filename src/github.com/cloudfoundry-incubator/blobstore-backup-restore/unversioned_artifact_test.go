package blobstore_test

import (
	"os"

	. "github.com/cloudfoundry-incubator/blobstore-backup-restore"

	"io/ioutil"

	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UnversionedFileArtifact", func() {
	var backupDir string
	var artifactPath string
	var fileArtifact UnversionedArtifact

	BeforeEach(func() {
		backupDir, _ = ioutil.TempDir("", "bbr_test_")
		artifactPath = filepath.Join(backupDir, "blobstore.json")
		fileArtifact = NewUnversionedFileArtifact(artifactPath)
	})

	AfterEach(func() {
		os.RemoveAll(backupDir)
	})

	It("saves the artifact to a file", func() {
		backup := map[string]BackupBucketAddress{
			"droplets": {
				BucketName:   "my_droplets_bucket",
				BucketRegion: "my_droplets_region",
				Path:         "my/droplets",
			},
			"buildpacks": {
				BucketName:   "my_buildpacks_bucket",
				BucketRegion: "my_buildpacks_region",
				Path:         "my/buildpacks",
			},
			"packages": {
				BucketName:   "my_packages_bucket",
				BucketRegion: "my_packages_region",
				Path:         "my/packages",
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
			fileArtifact = NewUnversionedFileArtifact("/this/path/does/not/exist")
		})

		It("returns an error", func() {
			err := fileArtifact.Save(map[string]BackupBucketAddress{})
			Expect(err).To(MatchError(ContainSubstring("could not write backup file")))
		})
	})

	Context("when reading the file fails", func() {
		BeforeEach(func() {
			fileArtifact = NewUnversionedFileArtifact("/this/path/does/not/exist")
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
