package gcs_test

import (
	"io/ioutil"
	"os"

	"github.com/cloudfoundry-incubator/gcs-blobstore-backup-restore"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Artifact", func() {
	var artifactFile *os.File
	var artifact gcs.Artifact

	Describe("Write", func() {
		Context("when the artifact file is writable", func() {
			BeforeEach(func() {
				var err error
				artifactFile, err = ioutil.TempFile("", "gcs_backup")
				Expect(err).NotTo(HaveOccurred())

				artifact = gcs.NewArtifact(artifactFile.Name())
			})

			It("persists the backups in it", func() {
				err := artifact.Write(map[string]gcs.BucketBackup{
					"bucket_identifier": {
						Name: "bucket_name",
						Blobs: []gcs.Blob{
							{Name: "a_blob", GenerationID: 123},
							{Name: "another_blob", GenerationID: 321},
						},
					},
				})

				Expect(err).NotTo(HaveOccurred())
				fileContent, err := ioutil.ReadFile(artifactFile.Name())
				Expect(err).NotTo(HaveOccurred())
				Expect(fileContent).To(MatchJSON(`{
					"bucket_identifier": {
						"name": "bucket_name",
						"blobs": [
							{"name": "a_blob", "generation_id": 123},
							{"name": "another_blob", "generation_id": 321}
						]
					}
				}`))
			})
		})

		Context("when the artifact file is not writable", func() {
			BeforeEach(func() {
				artifact = gcs.NewArtifact("")
			})

			It("returns an error", func() {
				err := artifact.Write(map[string]gcs.BucketBackup{})

				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Read", func() {
		Context("when the artifact file is readable", func() {
			It("reads the backups", func() {
				artifactFile, err := ioutil.TempFile("", "azure_restore_artifact")
				Expect(err).NotTo(HaveOccurred())
				err = ioutil.WriteFile(artifactFile.Name(), []byte(`{
					"bucket_identifier": {
						"name": "bucket_name",
						"blobs": [
							{"name": "a_blob", "generation_id": 123},
							{"name": "another_blob", "generation_id": 321}
						]
					}
				}`), 0644)
				Expect(err).NotTo(HaveOccurred())
				artifact = gcs.NewArtifact(artifactFile.Name())

				backups, err := artifact.Read()

				Expect(err).NotTo(HaveOccurred())
				Expect(backups).To(Equal(map[string]gcs.BucketBackup{
					"bucket_identifier": {
						Name: "bucket_name",
						Blobs: []gcs.Blob{
							{Name: "a_blob", GenerationID: 123},
							{Name: "another_blob", GenerationID: 321},
						},
					},
				}))
			})
		})

		Context("when the artifact file is not readable", func() {
			It("reports an error", func() {
				artifact = gcs.NewArtifact("does_not_exist")

				_, err := artifact.Read()

				Expect(err).To(HaveOccurred())
			})
		})

		Context("when the artifact file is not valid JSON", func() {
			It("reports an error", func() {
				artifactFile, err := ioutil.TempFile("", "azure_restore_artifact")
				Expect(err).NotTo(HaveOccurred())
				err = ioutil.WriteFile(artifactFile.Name(), []byte{}, 0644)
				Expect(err).NotTo(HaveOccurred())
				artifact = gcs.NewArtifact(artifactFile.Name())

				_, err = artifact.Read()

				Expect(err).To(HaveOccurred())
			})
		})
	})
})
