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
				err := restorer.Restore(map[string]azure.ContainerBackup{
					"first": {
						Name: firstContainerName,
						Blobs: []azure.Blob{
							{Name: "file_1_a", ETag: "1A"},
							{Name: "file_1_b", ETag: "1B"},
						},
					},
					"second": {
						Name: secondContainerName,
						Blobs: []azure.Blob{
							{Name: "file_2_a", ETag: "2A"},
						},
					},
				})

				Expect(err).NotTo(HaveOccurred())

				Expect(firstContainer.CopyFromCallCount()).To(Equal(2))

				containerName1a, blobName1a, etag1a := firstContainer.CopyFromArgsForCall(0)
				Expect(blobName1a).To(Equal("file_1_a"))
				Expect(etag1a).To(Equal("1A"))
				Expect(containerName1a).To(Equal(firstContainerName))

				containerName1b, blobName1b, etag1b := firstContainer.CopyFromArgsForCall(1)
				Expect(blobName1b).To(Equal("file_1_b"))
				Expect(etag1b).To(Equal("1B"))
				Expect(containerName1b).To(Equal(firstContainerName))

				Expect(secondContainer.CopyFromCallCount()).To(Equal(1))

				containerName2a, blobName2a, etag2a := secondContainer.CopyFromArgsForCall(0)
				Expect(blobName2a).To(Equal("file_2_a"))
				Expect(etag2a).To(Equal("2A"))
				Expect(containerName2a).To(Equal(secondContainerName))
			})
		})

		Context("when copying one of the blobs fails", func() {
			It("returns the error", func() {
				secondContainer.CopyFromReturns(errors.New("ooops"))

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

				Expect(firstContainer.CopyFromCallCount()).To(BeZero())
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
