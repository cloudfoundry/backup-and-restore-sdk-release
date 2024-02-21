package azure_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"azure-blobstore-backup-restore"
	"azure-blobstore-backup-restore/fakes"

	"github.com/Azure/azure-storage-blob-go/azblob"
)

var _ = Describe("Retry logic in SDKContainer for ListBlobs", func() {
	var fakeBlobLister *fakes.FakeBlobLister
	var sdkContainer azure.SDKContainer

	BeforeEach(func() {
		fakeBlobLister = &fakes.FakeBlobLister{}

		var err error
		sdkContainer, err = azure.NewTestSDKContainer("SSKS", azure.StorageAccount{}, azure.DefaultEnvironment, fakeBlobLister)
		Expect(err).To(BeNil())
	})

	When("we consistently get any error from the azure API", func() {

		BeforeEach(func() {
			fakeBlobLister.ListBlobsFlatSegmentReturns(nil, fmt.Errorf("Oh no!"))
		})
		It("retries three times and fails", func() {
			_, err := sdkContainer.ListBlobs()
			Expect(err).NotTo(BeNil())
			Expect(err).To(MatchError(ContainSubstring("[Oh no! Oh no! Oh no!]")), "The error should wrap the error from Azure")
			Expect(err).To(MatchError(ContainSubstring("failed listing blobs in container")), "The error should say what we were trying to do")
			Expect(fakeBlobLister.ListBlobsFlatSegmentCallCount()).To(Equal(3))
		})
	})

	When("azure api fails only twice", func() {

		BeforeEach(func() {
			fakeBlobLister.ListBlobsFlatSegmentReturnsOnCall(0, nil, fmt.Errorf("Oh no!"))
			fakeBlobLister.ListBlobsFlatSegmentReturnsOnCall(1, nil, fmt.Errorf("Oh no!"))
			fakeBlobLister.ListBlobsFlatSegmentReturnsOnCall(2, createGoodAzureListBlobsResponse(), nil)
		})
		It("then succeeds", func() {
			_, err := sdkContainer.ListBlobs()
			Expect(err).To(BeNil())
		})
	})

})

// createGoodAzureListBlobsResponse creates a minimal ListBlobsFlatSegmentResponse that represents a
// singleton blob list. This is the sort of thing that is returned by ListBlobsFlatSegment
// https://github.com/Azure/azure-storage-blob-go/blob/deef49049c414ab2c8a6370df8f2ae7f06ca4af5/azblob/zz_generated_models.go#L5382
func createGoodAzureListBlobsResponse() *azblob.ListBlobsFlatSegmentResponse {
	markerVal := ""
	goodResponse := &azblob.ListBlobsFlatSegmentResponse{
		Segment: azblob.BlobFlatListSegment{
			BlobItems: []azblob.BlobItemInternal{
				azblob.BlobItemInternal{
					Name: "My test blob",
					Properties: azblob.BlobPropertiesInternal{
						Etag: "My test ETag",
					},
				},
			},
		},
		NextMarker: azblob.Marker{
			Val: &markerVal,
		},
	}
	return goodResponse
}
