package azure_test

import (
	"io/ioutil"

	"os"

	"github.com/cloudfoundry-incubator/azure-blobstore-backup-restore"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Artifact", func() {
	var artifactFile *os.File
	var artifact azure.Artifact

	Describe("Write", func() {
		Context("when the artifact file is writable", func() {
			BeforeEach(func() {
				var err error
				artifactFile, err = ioutil.TempFile("", "azure_backup")
				Expect(err).NotTo(HaveOccurred())
				artifact = azure.NewArtifact(artifactFile.Name())
			})

			It("persists the backups in it", func() {
				err := artifact.Write(map[string]azure.ContainerBackup{
					"container_id": {
						Name: "container_name",
						Blobs: []azure.Blob{
							{Name: "a_blob", Hash: "abc123"},
							{Name: "another_blob", Hash: "def456"},
						},
					},
				})

				Expect(err).NotTo(HaveOccurred())

				fileContent, err := ioutil.ReadFile(artifactFile.Name())

				Expect(fileContent).To(MatchJSON(`{
					"container_id": {
						"name": "container_name",
						"blobs": [
							{ "name": "a_blob", "hash": "abc123" },
							{"name": "another_blob", "hash": "def456" }
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
})
