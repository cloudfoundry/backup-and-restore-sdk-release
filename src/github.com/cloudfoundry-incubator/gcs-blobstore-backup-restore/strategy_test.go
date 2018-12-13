package gcs_test

import (
	"errors"

	"sync/atomic"

	"github.com/cloudfoundry-incubator/gcs-blobstore-backup-restore"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Strategy", func() {
	Context("when running multiple actions in parallel", func() {
		It("calls the action on every blob", func() {
			strategy := gcs.NewParallelStrategy()
			blobs := []gcs.Blob{{}, {}, {}}
			var counter int32
			action := func(blob gcs.Blob) error {
				atomic.AddInt32(&counter, 1)
				return nil
			}

			errs := strategy.Run(blobs, action)

			Expect(counter).To(Equal(int32(3)))
			Expect(errs).To(BeEmpty())
		})
	})

	Context("when the actions fail", func() {
		It("returns the errors from each action", func() {
			strategy := gcs.NewParallelStrategy()
			blobs := []gcs.Blob{gcs.NewBlob("one"), gcs.NewBlob("two")}
			action := func(blob gcs.Blob) error {
				return errors.New(blob.Name())
			}

			errs := strategy.Run(blobs, action)

			Expect(errs).To(ConsistOf(MatchError("one"), MatchError("two")))
		})
	})
})
