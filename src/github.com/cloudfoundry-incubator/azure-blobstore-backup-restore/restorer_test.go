package azure_test

import (
	"errors"

	"github.com/cloudfoundry-incubator/azure-blobstore-backup-restore"
	"github.com/cloudfoundry-incubator/azure-blobstore-backup-restore/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Restorer", func() {
	var firstContainer *fakes.FakeContainer
	var secondContainer *fakes.FakeContainer

	var restorer azure.Restorer

	const firstContainerName = "first-container-name"
	const secondContainerName = "second-container-name"

	BeforeEach(func() {
		firstContainer = new(fakes.FakeContainer)
		secondContainer = new(fakes.FakeContainer)

		firstContainer.NameReturns(firstContainerName)
		secondContainer.NameReturns(secondContainerName)

		firstContainer.SoftDeleteEnabledReturns(true, nil)
		secondContainer.SoftDeleteEnabledReturns(true, nil)

		restorer = azure.NewRestorer(map[string]azure.Container{
			"first":  firstContainer,
			"second": secondContainer,
		})
	})

	Describe("Restore", func() {
		Context("when the artifact is valid", func() {
			It("rolls back each blob to the specified ETag", func() {
				firstContainerBlobs := []azure.Blob{
					{Name: "file_1_a", ETag: "1A"},
					{Name: "file_1_b", ETag: "1B"},
				}

				secondContainerBlobs := []azure.Blob{
					{Name: "file_2_a", ETag: "2A"},
				}

				err := restorer.Restore(map[string]azure.ContainerBackup{
					"first": {
						Name:  firstContainerName,
						Blobs: firstContainerBlobs,
					},
					"second": {
						Name:  secondContainerName,
						Blobs: secondContainerBlobs,
					},
				})

				Expect(err).NotTo(HaveOccurred())

				Expect(firstContainer.CopyBlobsFromCallCount()).To(Equal(1))

				actualContainerName, actualBlobsToCopy := firstContainer.CopyBlobsFromArgsForCall(0)
				Expect(actualContainerName).To(Equal(firstContainerName))
				Expect(actualBlobsToCopy).To(Equal(firstContainerBlobs))

				Expect(secondContainer.CopyBlobsFromCallCount()).To(Equal(1))

				actualContainerName, actualBlobsToCopy = secondContainer.CopyBlobsFromArgsForCall(0)
				Expect(actualContainerName).To(Equal(secondContainerName))
				Expect(actualBlobsToCopy).To(Equal(secondContainerBlobs))
			})
		})

		Context("when copying one of the containers fails", func() {
			It("returns the error", func() {
				secondContainer.CopyBlobsFromReturns(errors.New("ooops"))

				err := restorer.Restore(map[string]azure.ContainerBackup{
					"first": {
						Name: firstContainerName,
						Blobs: []azure.Blob{
							{Name: "file_1_a", ETag: "1A"},
						},
					},
					"second": {
						Name: secondContainerName,
						Blobs: []azure.Blob{
							{Name: "file_2_a", ETag: "2A"},
						},
					},
				})

				Expect(err).To(MatchError("ooops"))
			})
		})

		Context("when the container does not have soft delete enabled", func() {
			It("returns an error", func() {
				firstContainer.SoftDeleteEnabledReturns(false, nil)

				err := restorer.Restore(map[string]azure.ContainerBackup{
					"first": {
						Name:  firstContainerName,
						Blobs: []azure.Blob{},
					}})

				Expect(err).To(MatchError("soft delete is not enabled on the given storage account"))
			})
		})

		Context("when second container does not have soft delete enabled", func() {
			It("it does not copy blobs into the first container", func() {
				firstContainer.SoftDeleteEnabledReturns(true, nil)
				secondContainer.SoftDeleteEnabledReturns(false, nil)

				err := restorer.Restore(map[string]azure.ContainerBackup{
					"first": {
						Name:  firstContainerName,
						Blobs: []azure.Blob{{Name: "file_1_a", ETag: "1A"}},
					},
					"second": {
						Name:  secondContainerName,
						Blobs: []azure.Blob{}}})

				Expect(err).To(MatchError("soft delete is not enabled on the given storage account"))

				Expect(firstContainer.CopyBlobsFromCallCount()).To(BeZero())
			})
		})

		Context("when checking soft delete fails", func() {
			It("returns an error", func() {
				secondContainer.SoftDeleteEnabledReturns(false, errors.New("ooops"))

				err := restorer.Restore(map[string]azure.ContainerBackup{
					"second": {
						Name:  secondContainerName,
						Blobs: []azure.Blob{},
					}})

				Expect(err).To(MatchError("ooops"))
			})
		})
	})
})
