package azure_test

import (
	"os"

	"azure-blobstore-backup-restore"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Artifact", func() {
	var artifactFile *os.File
	var artifact azure.Artifact

	Describe("Write", func() {
		Context("when the artifact file is writable", func() {
			BeforeEach(func() {
				var err error
				artifactFile, err = os.CreateTemp("", "azure_backup")
				Expect(err).NotTo(HaveOccurred())
				artifact = azure.NewArtifact(artifactFile.Name())
			})

			It("persists the backups in it", func() {
				err := artifact.Write(map[string]azure.ContainerBackup{
					"container_id": {
						Name: "container_name",
						Blobs: []azure.BlobId{
							{Name: "a_blob", ETag: "abc123"},
							{Name: "another_blob", ETag: "def456"},
						},
					},
				})

				Expect(err).NotTo(HaveOccurred())

				fileContent, err := os.ReadFile(artifactFile.Name())

				Expect(fileContent).To(MatchJSON(`{
					"container_id": {
						"name": "container_name",
						"blobs": [
							{"name": "a_blob", "etag": "abc123"},
							{"name": "another_blob", "etag": "def456"}
						]
					}
				}`))
			})
		})

		Context("when the artifact file is not writable", func() {
			BeforeEach(func() {
				artifact = azure.NewArtifact("")
			})

			It("returns an error", func() {
				err := artifact.Write(map[string]azure.ContainerBackup{})

				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("Read", func() {
		Context("when the artifact file is readable", func() {
			It("reads the backups", func() {
				artifactFile, err := os.CreateTemp("", "azure_restore_artifact")
				Expect(err).NotTo(HaveOccurred())
				err = os.WriteFile(artifactFile.Name(), []byte(`{
					"container_id": {
						"name": "container_name",
						"blobs": [
							{"name": "a_blob", "etag": "abc123"},
							{"name": "another_blob", "etag": "def456"}
						]
					}
				}`), 0644)
				Expect(err).NotTo(HaveOccurred())

				artifact = azure.NewArtifact(artifactFile.Name())

				backups, err := artifact.Read()
				Expect(err).NotTo(HaveOccurred())
				Expect(backups).To(Equal(map[string]azure.ContainerBackup{
					"container_id": {
						Name: "container_name",
						Blobs: []azure.BlobId{
							{Name: "a_blob", ETag: "abc123"},
							{Name: "another_blob", ETag: "def456"},
						},
					},
				}))
			})
		})

		Context("when the artifact file is not readable", func() {
			It("reports an error", func() {
				artifact = azure.NewArtifact("does_not_exist")

				backups, err := artifact.Read()

				Expect(backups).To(BeNil())
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when the artifact file is not valid JSON", func() {
			It("reports an error", func() {
				artifactFile, err := os.CreateTemp("", "azure_restore_artifact")
				Expect(err).NotTo(HaveOccurred())
				err = os.WriteFile(artifactFile.Name(), []byte{}, 0644)
				Expect(err).NotTo(HaveOccurred())

				artifact = azure.NewArtifact(artifactFile.Name())

				backups, err := artifact.Read()

				Expect(backups).To(BeNil())
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
