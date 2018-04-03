package blobstore_test

import (
	"os"

	. "github.com/cloudfoundry-incubator/blobstore-backup-restore"

	"io/ioutil"

	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("VersionedFileArtifact", func() {
	var backupDir string
	var artifactPath string
	var fileArtifact VersionedFileArtifact

	BeforeEach(func() {
		backupDir, _ = ioutil.TempDir("", "bbr_test_")
		artifactPath = filepath.Join(backupDir, "blobstore.json")
		fileArtifact = NewVersionedFileArtifact(artifactPath)
	})

	AfterEach(func() {
		os.RemoveAll(backupDir)
	})

	It("saves the artifact to a file", func() {
		backup := map[string]BucketSnapshot{
			"droplets": {
				BucketName: "my_droplets_bucket",
				RegionName: "my_droplets_region",
				Versions: []BlobVersion{
					{BlobKey: "one", Id: "11"},
					{BlobKey: "two", Id: "21"},
				},
			},
			"buildpacks": {
				BucketName: "my_buildpacks_bucket",
				RegionName: "my_buildpacks_region",
				Versions: []BlobVersion{
					{BlobKey: "three", Id: "31"},
				},
			},
			"packages": {
				BucketName: "my_packages_bucket",
				RegionName: "my_packages_region",
				Versions: []BlobVersion{
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
			fileArtifact = NewVersionedFileArtifact("/this/path/does/not/exist")
		})

		It("returns an error", func() {
			err := fileArtifact.Save(map[string]BucketSnapshot{})
			Expect(err).To(MatchError(ContainSubstring("could not write backup file")))
		})
	})

	Context("when reading the file fails", func() {
		BeforeEach(func() {
			fileArtifact = NewVersionedFileArtifact("/this/path/does/not/exist")
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
