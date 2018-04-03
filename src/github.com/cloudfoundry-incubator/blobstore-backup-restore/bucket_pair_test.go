package blobstore_test

import (
	. "github.com/cloudfoundry-incubator/blobstore-backup-restore"
	"github.com/cloudfoundry-incubator/blobstore-backup-restore/fakes"

	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Backup", func() {

	var (
		liveBucket   *fakes.FakeUnversionedBucket
		backupBucket *fakes.FakeUnversionedBucket
		bucketPair   S3BucketPair
		address      BackupBucketAddress
		err          error
	)

	BeforeEach(func() {
		liveBucket = new(fakes.FakeUnversionedBucket)
		backupBucket = new(fakes.FakeUnversionedBucket)
		bucketPair = S3BucketPair{
			LiveBucket:   liveBucket,
			BackupBucket: backupBucket}

		liveBucket.ListFilesReturns([]string{"path1/file1", "path2/file2"}, nil)
		liveBucket.NameReturns("liveBucket")
		liveBucket.RegionNameReturns("liveBucketRegion")
		backupBucket.NameReturns("backupBucket")
		backupBucket.RegionNameReturns("backupBucketRegion")
	})

	JustBeforeEach(func() {
		address, err = bucketPair.Backup("destination-string")
	})

	It("copies all the files in the bucket", func() {
		By("not failing", func() {
			Expect(err).NotTo(HaveOccurred())
		})
		By("Listing the files in the bucket", func() {
			Expect(liveBucket.ListFilesCallCount()).To(Equal(1))
		})

		By("calling copy for each file in the bucket", func() {
			Expect(backupBucket.CopyCallCount()).To(Equal(2))
			expectedKey, expectedDestinationPath, expectedOriginBucketName, expectedOriginBucketRegion := backupBucket.CopyArgsForCall(0)
			Expect(expectedKey).To(Equal("path1/file1"))
			Expect(expectedDestinationPath).To(Equal("destination-string"))
			Expect(expectedOriginBucketName).To(Equal("liveBucket"))
			Expect(expectedOriginBucketRegion).To(Equal("liveBucketRegion"))
			expectedKey, expectedDestinationPath, expectedOriginBucketName, expectedOriginBucketRegion = backupBucket.CopyArgsForCall(1)
			Expect(expectedKey).To(Equal("path2/file2"))
			Expect(expectedDestinationPath).To(Equal("destination-string"))
			Expect(expectedOriginBucketName).To(Equal("liveBucket"))
			Expect(expectedOriginBucketRegion).To(Equal("liveBucketRegion"))
		})

		By("returning the address of the backup bucket", func() {
			Expect(address).To(Equal(BackupBucketAddress{
				BucketName:   "backupBucket",
				BucketRegion: "backupBucketRegion",
				Path:         "destination-string",
			}))
		})
	})

	Context("when ListFiles fails", func() {
		BeforeEach(func() {
			liveBucket.ListFilesReturns([]string{}, fmt.Errorf("cannot list files"))
		})

		It("should fail", func() {
			Expect(err).To(MatchError("cannot list files"))
		})
	})

	Context("when Copy fails", func() {
		BeforeEach(func() {
			backupBucket.CopyReturns(fmt.Errorf("cannot copy file"))
		})

		It("should fail", func() {
			Expect(err).To(MatchError("cannot copy file"))
		})
	})
})
