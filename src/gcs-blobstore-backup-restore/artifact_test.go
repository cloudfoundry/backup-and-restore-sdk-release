package gcs_test

import (
	"fmt"
	"io/ioutil"
	"os"

	"gcs-blobstore-backup-restore"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Artifact", func() {
	Describe("Write", func() {
		var artifactFile *os.File
		var artifact gcs.Artifact

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
					"bucket_identifier_2": {
						SameBucketAs: "bucket_identifier",
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
					},
					"bucket_identifier_2": {
						"bucket_name": "",
						"path": "",
						"same_bucket_as": "bucket_identifier"
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
		Context("when the artifact file exists", func() {
			Context("and is valid json", func() {
				It("returns the artifact contents", func() {
					artifactFile, err := ioutil.TempFile("", "gcs_backup")
					Expect(err).NotTo(HaveOccurred())

					_, err = artifactFile.Write([]byte(`{
					"bucket_identifier": {
						"bucket_name": "bucket_name",
						"path": "a_path"
					},
					"bucket_identifier_2": {
						"bucket_name": "",
						"path": "",
						"same_bucket_as": "bucket_identifier"
					}
				}`))
					Expect(err).NotTo(HaveOccurred())

					artifact := gcs.NewArtifact(artifactFile.Name())
					content, err := artifact.Read()
					Expect(err).NotTo(HaveOccurred())
					Expect(content).To(Equal(map[string]gcs.BucketBackup{
						"bucket_identifier": {
							BucketName: "bucket_name",
							Path:       "a_path",
						},
						"bucket_identifier_2": {
							SameBucketAs: "bucket_identifier",
						},
					}))
				})
			})

			Context("and is not valid json", func() {
				It("returns an error", func() {
					artifactFile, err := ioutil.TempFile("", "gcs_backup")
					Expect(err).NotTo(HaveOccurred())

					_, err = artifactFile.Write([]byte(`not-valid{json`))
					Expect(err).NotTo(HaveOccurred())

					artifact := gcs.NewArtifact(artifactFile.Name())
					_, err = artifact.Read()
					Expect(err).To(MatchError(ContainSubstring(fmt.Sprintf("Failed to unmarshall artifact %s", artifactFile.Name()))))

				})
			})
		})

		Context("when the artifact file does not exist", func() {
			It("returns an error", func() {
				artifact := gcs.NewArtifact("/tmp/foo/bar/i/dont/exist")
				_, err := artifact.Read()
				Expect(err).To(MatchError(ContainSubstring("Failed to read artifact file /tmp/foo/bar/i/dont/exist")))
			})
		})
	})
})
