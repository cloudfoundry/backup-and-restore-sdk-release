package azure_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"azure-blobstore-backup-restore"
	"azure-blobstore-backup-restore/fakes"
)

var _ = Describe("Retry logic in SDKContainer for ListBlobs", func() {
	When("we get any error from the azure API", func() {
		It("retries three times", func() {
			fakeBlobLister := &fakes.FakeBlobLister{}
			fakeBlobLister.ListBlobsFlatSegmentReturns(nil, fmt.Errorf("Oh no!"))

			sdkContainer, err := azure.NewTestSDKContainer("SSKS", azure.StorageAccount{}, azure.DefaultEnvironment, fakeBlobLister)
			Expect(err).To(BeNil())

			_, err = sdkContainer.ListBlobs()
			Expect(err).NotTo(BeNil())
			Expect(err).To(MatchError(ContainSubstring("[Oh no! Oh no! Oh no!]")))
			Expect(fakeBlobLister.ListBlobsFlatSegmentCallCount()).To(Equal(3))
		})
	})
})
