package gcs_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"errors"

	"fmt"

	"github.com/cloudfoundry-incubator/gcs-blobstore-backup-restore"
	"github.com/cloudfoundry-incubator/gcs-blobstore-backup-restore/fakes"
)

var _ = Describe("Backuper", func() {
	var firstBucket *fakes.FakeBucket
	var secondBucket *fakes.FakeBucket
	var thirdBucket *fakes.FakeBucket

	var backuper gcs.Backuper

	const firstBucketName = "first-bucket-name"
	const secondBucketName = "second-bucket-name"
	const thirdBucketName = "third-bucket-name"

	BeforeEach(func() {
		firstBucket = new(fakes.FakeBucket)
		secondBucket = new(fakes.FakeBucket)
		thirdBucket = new(fakes.FakeBucket)

		firstBucket.NameReturns(firstBucketName)
		secondBucket.NameReturns(secondBucketName)
		thirdBucket.NameReturns(thirdBucketName)

		firstBucket.VersioningEnabledReturns(true, nil)
		secondBucket.VersioningEnabledReturns(true, nil)
		thirdBucket.VersioningEnabledReturns(true, nil)

		backuper = gcs.NewBackuper(map[string]gcs.Bucket{
			"first":  firstBucket,
			"second": secondBucket,
			"third":  thirdBucket,
		})
	})

	Context("when fetching the blobs succeeds", func() {
		It("returns a map of all fetched blobs for each container", func() {
			firstBucket.ListBlobsReturns([]gcs.Blob{
				{Name: "file_1_a"},
				{Name: "file_1_b"},
			}, nil)
			secondBucket.ListBlobsReturns([]gcs.Blob{}, nil)
			thirdBucket.ListBlobsReturns([]gcs.Blob{
				{Name: "file_3_a"},
			}, nil)

			backups, err := backuper.Backup()

			Expect(err).NotTo(HaveOccurred())
			Expect(backups).To(Equal(map[string]gcs.BucketBackup{
				"first": {
					Name: firstBucketName,
					Blobs: []gcs.Blob{
						{Name: "file_1_a"},
						{Name: "file_1_b"},
					},
				},
				"second": {
					Name:  secondBucketName,
					Blobs: []gcs.Blob{},
				},
				"third": {
					Name: thirdBucketName,
					Blobs: []gcs.Blob{
						{Name: "file_3_a"},
					},
				},
			}))
		})
	})

	Context("when fetching the blobs from one of the containers fails", func() {
		It("returns the error", func() {
			thirdBucket.ListBlobsReturns(nil, errors.New("ooops"))

			_, err := backuper.Backup()

			Expect(thirdBucket.ListBlobsCallCount()).To(Equal(1))
			Expect(err).To(MatchError("ooops"))
		})
	})

	Context("when one of the buckets does not have versioning enabled", func() {
		It("returns an error", func() {
			secondBucket.VersioningEnabledReturns(false, nil)

			_, err := backuper.Backup()

			Expect(secondBucket.VersioningEnabledCallCount()).To(Equal(1))
			Expect(firstBucket.ListBlobsCallCount()).To(BeZero())
			Expect(secondBucket.ListBlobsCallCount()).To(BeZero())
			Expect(thirdBucket.ListBlobsCallCount()).To(BeZero())
			Expect(err).To(MatchError(fmt.Sprintf("versioning is not enabled on bucket: %s", secondBucketName)))
		})
	})

	Context("when checking soft delete fails", func() {
		It("returns an error", func() {
			secondBucket.VersioningEnabledReturns(false, errors.New("ooops"))

			_, err := backuper.Backup()

			Expect(err).To(MatchError("ooops"))
		})
	})
})
