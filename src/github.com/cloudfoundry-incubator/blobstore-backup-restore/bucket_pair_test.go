package blobstore_test

import (
	. "github.com/cloudfoundry-incubator/blobstore-backup-restore"
	"github.com/cloudfoundry-incubator/blobstore-backup-restore/s3/fakes"

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

		liveBucket.NameReturns("liveBucket")
		liveBucket.RegionNameReturns("liveBucketRegion")
		backupBucket.NameReturns("backupBucket")
		backupBucket.RegionNameReturns("backupBucketRegion")
	})

	JustBeforeEach(func() {
		address, err = bucketPair.Backup("destination-string")
	})

	Context("When there are files in the bucket", func() {

		BeforeEach(func() {
			liveBucket.ListFilesReturns([]string{"path1/file1", "path2/file2"}, nil)
		})

		It("copies all the files in the bucket", func() {
			By("succeeding", func() {
				Expect(err).NotTo(HaveOccurred())
			})
			By("Listing the files in the bucket", func() {
				Expect(liveBucket.ListFilesCallCount()).To(Equal(1))
			})

			By("calling copy for each file in the bucket", func() {
				Expect(backupBucket.CopyObjectCallCount()).To(Equal(2))
				expectedKey, expectedOriginPath, expectedDestinationPath, expectedOriginBucketName, expectedOriginBucketRegion :=
					backupBucket.CopyObjectArgsForCall(0)
				Expect(expectedKey).To(Equal("path1/file1"))
				Expect(expectedOriginPath).To(Equal(""))
				Expect(expectedDestinationPath).To(Equal("destination-string"))
				Expect(expectedOriginBucketName).To(Equal("liveBucket"))
				Expect(expectedOriginBucketRegion).To(Equal("liveBucketRegion"))
				expectedKey, expectedOriginPath, expectedDestinationPath, expectedOriginBucketName, expectedOriginBucketRegion =
					backupBucket.CopyObjectArgsForCall(1)
				Expect(expectedKey).To(Equal("path2/file2"))
				Expect(expectedOriginPath).To(Equal(""))
				Expect(expectedDestinationPath).To(Equal("destination-string"))
				Expect(expectedOriginBucketName).To(Equal("liveBucket"))
				Expect(expectedOriginBucketRegion).To(Equal("liveBucketRegion"))
			})

			By("returning the address of the backup bucket", func() {
				Expect(address).To(Equal(BackupBucketAddress{
					BucketName:   "backupBucket",
					BucketRegion: "backupBucketRegion",
					Path:         "destination-string",
					EmptyBackup:  false,
				}))
			})
		})

		Context("when CopyObject fails", func() {
			BeforeEach(func() {
				backupBucket.CopyObjectReturns(fmt.Errorf("cannot copy file"))
			})

			It("should fail", func() {
				Expect(err).To(MatchError("cannot copy file"))
			})
		})
	})

	Context("When there are no files in the bucket", func() {
		BeforeEach(func() {
			liveBucket.ListFilesReturns([]string{}, nil)
		})

		It("Records that information in the backup artifact", func() {

			By("succeeding", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			By("not calling copy", func() {
				Expect(backupBucket.CopyObjectCallCount()).To(Equal(0))
			})

			By("recording that the backup was empty", func() {
				Expect(address).To(Equal(BackupBucketAddress{
					BucketName:   "backupBucket",
					BucketRegion: "backupBucketRegion",
					Path:         "destination-string",
					EmptyBackup:  true,
				}))
			})
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

})

var _ = Describe("Restore", func() {
	var (
		backupBucket *fakes.FakeUnversionedBucket
		liveBucket   *fakes.FakeUnversionedBucket
		pair         UnversionedBucketPair
		err          error
	)

	JustBeforeEach(func() {
		err = pair.Restore("2015-12-13-05-06-07/my_bucket")
	})

	BeforeEach(func() {
		backupBucket = new(fakes.FakeUnversionedBucket)
		backupBucket.NameReturns("backupBucket")
		backupBucket.RegionNameReturns("backupRegion")

		liveBucket = new(fakes.FakeUnversionedBucket)

		backupBucket.ListFilesReturns([]string{"my_key"}, nil)

		pair = S3BucketPair{
			LiveBucket:   liveBucket,
			BackupBucket: backupBucket,
		}
	})

	It("successfully copies from the backup bucket to the live bucket", func() {
		By("not returning an error", func() {
			Expect(err).NotTo(HaveOccurred())
		})

		By("copying from the backup location to the live location", func() {
			Expect(backupBucket.ListFilesCallCount()).To(Equal(1))
			Expect(backupBucket.ListFilesArgsForCall(0)).To(Equal("2015-12-13-05-06-07/my_bucket"))
			Expect(liveBucket.CopyObjectCallCount()).To(Equal(1))
			key, originPath, destinationPath, originBucketName, originBucketRegion := liveBucket.CopyObjectArgsForCall(0)
			Expect(key).To(Equal("my_key"))
			Expect(originPath).To(Equal("2015-12-13-05-06-07/my_bucket"))
			Expect(destinationPath).To(Equal(""))
			Expect(originBucketName).To(Equal(backupBucket.Name()))
			Expect(originBucketRegion).To(Equal(backupBucket.RegionName()))
		})
	})

	Context("When ListFiles errors", func() {
		BeforeEach(func() {
			backupBucket.ListFilesReturns([]string{}, fmt.Errorf("cannot list files"))
		})

		It("errors", func() {
			Expect(err).To(MatchError(ContainSubstring("cannot list files")))
		})
	})

	Context("When CopyObject errors", func() {
		BeforeEach(func() {
			liveBucket.CopyObjectReturns(fmt.Errorf("cannot copy object"))
		})

		It("errors", func() {
			Expect(err).To(MatchError(ContainSubstring("cannot copy object")))
		})
	})

	Context("When listFiles returns no files", func() {
		BeforeEach(func() {
			backupBucket.ListFilesReturns([]string{}, nil)
		})

		It("errors saying there should be a backup", func() {
			Expect(err).To(MatchError(ContainSubstring("no files found in 2015-12-13-05-06-07/my_bucket " +
				"in bucket backupBucket to restore")))
		})
	})
})
