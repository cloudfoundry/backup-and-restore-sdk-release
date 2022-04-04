package azure_test

import (
	"errors"

	"github.com/cloudfoundry/backup-and-restore-sdk-release/src/azure-blobstore-backup-restore"
	"github.com/cloudfoundry/backup-and-restore-sdk-release/src/azure-blobstore-backup-restore/fakes"
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
		}, map[string]azure.StorageAccount{})
	})

	Describe("Restore", func() {
		Context("when the artifact is valid", func() {
			It("rolls back each blob to the specified ETag", func() {
				firstContainerBlobs := []azure.BlobId{
					{Name: "file_1_a", ETag: "1A"},
					{Name: "file_1_b", ETag: "1B"},
				}

				secondContainerBlobs := []azure.BlobId{
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

				Expect(firstContainer.CopyBlobsFromSameStorageAccountCallCount()).To(Equal(1))

				actualContainerName, actualBlobsToCopy := firstContainer.CopyBlobsFromSameStorageAccountArgsForCall(0)
				Expect(actualContainerName).To(Equal(firstContainerName))
				Expect(actualBlobsToCopy).To(Equal(firstContainerBlobs))

				Expect(secondContainer.CopyBlobsFromSameStorageAccountCallCount()).To(Equal(1))

				actualContainerName, actualBlobsToCopy = secondContainer.CopyBlobsFromSameStorageAccountArgsForCall(0)
				Expect(actualContainerName).To(Equal(secondContainerName))
				Expect(actualBlobsToCopy).To(Equal(secondContainerBlobs))
			})
		})

		Context("when copying one of the containers fails", func() {
			It("returns the error", func() {
				secondContainer.CopyBlobsFromSameStorageAccountReturns(errors.New("ooops"))

				err := restorer.Restore(map[string]azure.ContainerBackup{
					"first": {
						Name: firstContainerName,
						Blobs: []azure.BlobId{
							{Name: "file_1_a", ETag: "1A"},
						},
					},
					"second": {
						Name: secondContainerName,
						Blobs: []azure.BlobId{
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

				restorer = azure.NewRestorer(map[string]azure.Container{
					"first": firstContainer,
				}, map[string]azure.StorageAccount{})

				err := restorer.Restore(map[string]azure.ContainerBackup{
					"first": {
						Name:  firstContainerName,
						Blobs: []azure.BlobId{},
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
						Blobs: []azure.BlobId{{Name: "file_1_a", ETag: "1A"}},
					},
					"second": {
						Name:  secondContainerName,
						Blobs: []azure.BlobId{}}})

				Expect(err).To(MatchError("soft delete is not enabled on the given storage account"))

				Expect(firstContainer.CopyBlobsFromSameStorageAccountCallCount()).To(BeZero())
			})
		})

		Context("when checking soft delete fails", func() {
			It("returns an error", func() {
				secondContainer.SoftDeleteEnabledReturns(false, errors.New("ooops"))

				restorer = azure.NewRestorer(map[string]azure.Container{
					"second": secondContainer,
				}, map[string]azure.StorageAccount{})

				err := restorer.Restore(map[string]azure.ContainerBackup{
					"second": {
						Name:  secondContainerName,
						Blobs: []azure.BlobId{},
					}})

				Expect(err).To(MatchError("ooops"))
			})
		})

		Context("when the destination container belongs to a different storage account", func() {
			It("performs the copy using the specified storage account", func() {
				sourceStorageAccount := azure.StorageAccount{
					Name: "source-storage-account",
					Key:  "source-storage-key",
				}

				restorer = azure.NewRestorer(map[string]azure.Container{
					"container": firstContainer,
				}, map[string]azure.StorageAccount{
					"container": sourceStorageAccount,
				})

				blobs := []azure.BlobId{
					{Name: "file_1_a", ETag: "1A"},
					{Name: "file_1_b", ETag: "1B"},
				}

				err := restorer.Restore(map[string]azure.ContainerBackup{
					"container": {
						Name:  "source-container-name",
						Blobs: blobs,
					},
				})

				Expect(err).NotTo(HaveOccurred())

				Expect(firstContainer.CopyBlobsFromDifferentStorageAccountCallCount()).To(Equal(1))

				actualStorageAccount, actualContainerName, actualBlobsToCopy := firstContainer.CopyBlobsFromDifferentStorageAccountArgsForCall(0)
				Expect(actualStorageAccount).To(Equal(sourceStorageAccount))
				Expect(actualContainerName).To(Equal("source-container-name"))
				Expect(actualBlobsToCopy).To(Equal(blobs))
			})
		})

		Context("when there is a container referenced in the artifact that is not in the restore config", func() {
			BeforeEach(func() {
				restorer = azure.NewRestorer(map[string]azure.Container{
					"container": firstContainer,
				}, map[string]azure.StorageAccount{})
			})

			It("returns a useful error", func() {
				err := restorer.Restore(map[string]azure.ContainerBackup{
					"container": {
						Name:  firstContainerName,
						Blobs: []azure.BlobId{},
					},
					"not-in-restore-config": {
						Name:  "not-there",
						Blobs: []azure.BlobId{},
					},
				})

				Expect(err).To(MatchError("container not-in-restore-config is not mentioned in the restore config" +
					" but is present in the artifact"))
			})
		})

		Context("when there is a container specified in restore config but not in the backup artifact", func() {
			BeforeEach(func() {
				restorer = azure.NewRestorer(map[string]azure.Container{
					"container":  firstContainer,
					"container2": secondContainer,
				}, map[string]azure.StorageAccount{})
			})

			It("returns a useful error", func() {
				err := restorer.Restore(map[string]azure.ContainerBackup{
					"container": {
						Name:  firstContainerName,
						Blobs: []azure.BlobId{},
					},
				})

				Expect(err).To(MatchError("container container2 is mentioned in the restore config" +
					" but is not recorded in the artifact"))
			})
		})
	})
})
