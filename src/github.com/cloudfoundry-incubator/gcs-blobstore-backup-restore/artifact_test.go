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
						BucketName: "bucket_name",
						Path:       "a_path",
					},
				},
				)

				Expect(err).NotTo(HaveOccurred())
				fileContent, err := ioutil.ReadFile(artifactFile.Name())
				Expect(err).NotTo(HaveOccurred())
				Expect(fileContent).To(MatchJSON(`{
					"bucket_identifier": {
						"bucket_name": "bucket_name",
						"path": "a_path"
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
})
