package incremental_test

import (
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/incremental"
)

var _ = Describe("Artifact", func() {
	var artifactFile *os.File
	var artifact incremental.Artifact

	BeforeEach(func() {
		var err error
		artifactFile, err = ioutil.TempFile("", "s3_backup")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		err := os.Remove(artifactFile.Name())
		Expect(err).NotTo(HaveOccurred())
	})

	Context("when the backup artifact files exists", func() {
		It("loads the backupArtifact", func() {
			body := `{
			"bucket_id": {
				"bucket_name": "backup-bucket",
				"src_backup_directory_path": "2000_01_02_03_04_05/bucket_id",
				"blobs": ["2000_01_02_03_04_05/bucket_id/f0/fd/blob1/uuid", "2000_01_02_03_04_05/bucket_id/f0/fd/blob2/uuid"]
			}
		}`

			err := ioutil.WriteFile(artifactFile.Name(), []byte(body), 644)
			Expect(err).NotTo(HaveOccurred())

			artifact := incremental.NewArtifact(artifactFile.Name())
			bucketBackup, err := artifact.Load()
			Expect(err).NotTo(HaveOccurred())

			Expect(bucketBackup).To(Equal(map[string]incremental.Backup{
				"bucket_id": {
					BucketName: "backup-bucket",
					Blobs: []string{
						"2000_01_02_03_04_05/bucket_id/f0/fd/blob1/uuid",
						"2000_01_02_03_04_05/bucket_id/f0/fd/blob2/uuid",
					},
					SrcBackupDirectoryPath: "2000_01_02_03_04_05/bucket_id",
				},
			}))
		})
	})

	Context("when the back artifact file does not exist", func() {
		It("errors", func() {
			artifact := incremental.NewArtifact("does-not-exist-file")

			_, err := artifact.Load()

			Expect(err).To(MatchError(ContainSubstring("no such file or directory")))
		})
	})

	Context("when the back artifact file is not valid json", func() {
		It("errors", func() {
			err := ioutil.WriteFile(artifactFile.Name(), []byte("not-valid-json"), 644)
			Expect(err).NotTo(HaveOccurred())

			artifact := incremental.NewArtifact(artifactFile.Name())

			_, err = artifact.Load()

			Expect(err).To(MatchError(ContainSubstring("invalid character")))
		})
	})

	It("writes the backupArtifact", func() {
		artifact = incremental.NewArtifact(artifactFile.Name())

		backup := map[string]incremental.Backup{
			"bucket_id": {
				BucketName:   "backup-bucket",
				BucketRegion: "backup-bucket-region",
				Blobs: []string{
					"2000_01_02_03_04_05/bucket_id/f0/fd/blob1/uuid",
					"2000_01_02_03_04_05/bucket_id/f0/fd/blob2/uuid",
				},
				SrcBackupDirectoryPath: "2000_01_02_03_04_05/bucket_id",
			},
		}
		err := artifact.Write(backup)

		Expect(err).NotTo(HaveOccurred())

		fileContent, err := ioutil.ReadFile(artifactFile.Name())
		Expect(fileContent).To(MatchJSON(`{
			"bucket_id": {
				"bucket_name": "backup-bucket",
				"bucket_region": "backup-bucket-region",
				"src_backup_directory_path": "2000_01_02_03_04_05/bucket_id",
				"blobs": ["2000_01_02_03_04_05/bucket_id/f0/fd/blob1/uuid", "2000_01_02_03_04_05/bucket_id/f0/fd/blob2/uuid"]
			}
		}`))

		savedBackup, err := artifact.Load()
		Expect(err).NotTo(HaveOccurred())
		Expect(savedBackup).To(Equal(backup))
	})

	Context("when saving the file fails", func() {
		BeforeEach(func() {
			artifact = incremental.NewArtifact("/this/path/does/not/exist")
		})

		It("returns an error", func() {
			err := artifact.Write(map[string]incremental.Backup{})
			Expect(err).To(MatchError(ContainSubstring("could not write backup file")))
		})
	})

	Context("when reading the file fails", func() {
		BeforeEach(func() {
			artifact = incremental.NewArtifact("/this/path/does/not/exist")
		})

		It("returns an error", func() {
			backup, err := artifact.Load()
			Expect(backup).To(BeNil())
			Expect(err).To(MatchError(ContainSubstring("could not read backup file")))
		})
	})

	Context("when the artifact has an invalid format", func() {
		BeforeEach(func() {
			artifact = incremental.NewArtifact(artifactFile.Name())
			err := ioutil.WriteFile(artifactFile.Name(), []byte("THIS IS NOT VALID JSON"), 0666)
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns an error", func() {
			backup, err := artifact.Load()
			Expect(backup).To(BeNil())
			Expect(err).To(MatchError(ContainSubstring("backup file has an invalid format")))
		})
	})
})
