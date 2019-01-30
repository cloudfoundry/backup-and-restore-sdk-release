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
	BeforeEach(func() {
		var err error
		artifactFile, err = ioutil.TempFile("", "s3_backup")
		Expect(err).NotTo(HaveOccurred())

	})

	It("writes the artifact", func() {
		artifact := incremental.NewArtifact(artifactFile.Name())

		err := artifact.Write(map[string]incremental.BucketBackup{
			"bucket_id": {
				BucketName: "backup-bucket",
				Blobs: []string{
					"2000_01_02_03_04_05/bucket_id/f0/fd/blob1/uuid",
					"2000_01_02_03_04_05/bucket_id/f0/fd/blob2/uuid",
				},
				BackupDirectoryPath: "2000_01_02_03_04_05/bucket_id",
			},
		})

		Expect(err).NotTo(HaveOccurred())

		fileContent, err := ioutil.ReadFile(artifactFile.Name())
		Expect(fileContent).To(MatchJSON(`{
			"bucket_id": {
				"bucket_name": "backup-bucket",
				"backup_directory_path": "2000_01_02_03_04_05/bucket_id",
				"blobs": ["2000_01_02_03_04_05/bucket_id/f0/fd/blob1/uuid", "2000_01_02_03_04_05/bucket_id/f0/fd/blob2/uuid"]
			}
		}`))
	})
})
