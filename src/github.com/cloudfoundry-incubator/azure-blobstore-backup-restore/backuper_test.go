package azure_test

import (
	"errors"

	"fmt"

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

	const firstContainerName = "first-container-name"
	const secondContainerName = "second-container-name"
	const thirdContainerName = "third-container-name"

	BeforeEach(func() {
		firstContainer = new(fakes.FakeContainer)
		secondContainer = new(fakes.FakeContainer)
		thirdContainer = new(fakes.FakeContainer)

		firstContainer.NameReturns(firstContainerName)
		secondContainer.NameReturns(secondContainerName)
		thirdContainer.NameReturns(thirdContainerName)

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
						Name: firstContainerName,
						Blobs: []azure.Blob{
							{Name: "file_1_a", Hash: "1A"},
							{Name: "file_1_b", Hash: "1B"},
						},
					},
					"second": {
						Name:  secondContainerName,
						Blobs: []azure.Blob{},
					},
					"third": {
						Name: thirdContainerName,
						Blobs: []azure.Blob{
							{Name: "file_3_a", Hash: "3A"},
						},
					},
				}))
			})
		})

		Context("when one of the containers does not have soft delete enabled", func() {
			It("returns an error", func() {
				secondContainer.SoftDeleteIsDisabledReturns(true, nil)

				_, err := backuper.Backup()

				message := fmt.Sprintf("soft delete is not enabled on container: '%s'", secondContainerName)
				Expect(err).To(MatchError(message))
			})
		})

		Context("when checking soft delete fails", func() {
			It("returns an error", func() {
				secondContainer.SoftDeleteIsDisabledReturns(false, errors.New("ooops"))

				_, err := backuper.Backup()

				Expect(err).To(MatchError("ooops"))
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
