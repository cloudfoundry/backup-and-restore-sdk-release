package azure_test

import (
	"errors"

	"github.com/cloudfoundry-incubator/azure-blobstore-backup-restore"
	"github.com/cloudfoundry-incubator/azure-blobstore-backup-restore/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Backuper", func() {
	var firstContainer *fakes.FakeContainer
	var secondContainer *fakes.FakeContainer
	var thirdContainer *fakes.FakeContainer

	var backuper azure.Backuper

	BeforeEach(func() {
		firstContainer = new(fakes.FakeContainer)
		secondContainer = new(fakes.FakeContainer)
		thirdContainer = new(fakes.FakeContainer)

		firstContainer.NameReturns("first-container-name")
		secondContainer.NameReturns("second-container-name")
		thirdContainer.NameReturns("third-container-name")

		backuper = azure.NewBackuper(map[string]azure.Container{
			"first":  firstContainer,
			"second": secondContainer,
			"third":  thirdContainer,
		})
	})

	Describe("Backup", func() {
		Context("when fetching the blobs succeeds", func() {
			It("returns a map of all fetched blobs for each container", func() {
				firstContainer.ListBlobsReturns([]azure.Blob{
					{Name: "file_1_a", Hash: "1A"},
					{Name: "file_1_b", Hash: "1B"},
				}, nil)
				secondContainer.ListBlobsReturns([]azure.Blob{}, nil)
				thirdContainer.ListBlobsReturns([]azure.Blob{
					{Name: "file_3_a", Hash: "3A"},
				}, nil)

				backups, err := backuper.Backup()

				Expect(err).NotTo(HaveOccurred())
				Expect(backups).To(Equal(map[string]azure.ContainerBackup{
					"first": {
						Name: "first-container-name",
						Blobs: []azure.Blob{
							{Name: "file_1_a", Hash: "1A"},
							{Name: "file_1_b", Hash: "1B"},
						},
					},
					"second": {
						Name:  "second-container-name",
						Blobs: []azure.Blob{},
					},
					"third": {
						Name: "third-container-name",
						Blobs: []azure.Blob{
							{Name: "file_3_a", Hash: "3A"},
						},
					},
				}))
			})
		})

		Context("when fetching the blobs from one of the containers fails", func() {
			It("returns the error", func() {
				secondContainer.ListBlobsReturns(nil, errors.New("ooops"))

				backups, err := backuper.Backup()

				Expect(err).To(MatchError("ooops"))
				Expect(backups).To(BeNil())
			})
		})
	})
})
