package incremental_test

import (
	"github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/incremental"
	"github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/incremental/fakes"

	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("BucketPair", func() {
	var (
		liveBucket         *fakes.FakeBucket
		backupBucket       *fakes.FakeBucket
		bucketPair         incremental.RestoreBucketPair
		bucketBackup       incremental.BucketBackup
		err                error
		configLiveBucket   string
		configLiveRegion   string
		configBackupBucket string
		configBackupRegion string
	)

	BeforeEach(func() {
		configLiveBucket = "config_live_bucket"
		configLiveRegion = "config_live_region"
		configBackupBucket = "config_backup_bucket"
		configBackupRegion = "config_backup_region"

		liveBucket = new(fakes.FakeBucket)
		backupBucket = new(fakes.FakeBucket)
		bucketPair = incremental.NewIncrementalBucketPair(liveBucket, backupBucket)

		liveBucket.NameReturns(configLiveBucket)
		liveBucket.RegionReturns(configLiveRegion)
		backupBucket.NameReturns(configBackupBucket)
		backupBucket.RegionReturns(configBackupRegion)
	})

	//Describe("Backup", func() {
	//	JustBeforeEach(func() {
	//		bucketBackup, err = bucketPair.Backup("destination-string")
	//	})
	//
	//	Context("When there are files in the bucket", func() {
	//		BeforeEach(func() {
	//			liveBucket.ListFilesReturns([]string{"path1/file1", "path2/file2"}, nil)
	//		})
	//
	//		It("copies all the files in the bucket", func() {
	//			By("succeeding", func() {
	//				Expect(err).NotTo(HaveOccurred())
	//			})
	//			By("Listing the files in the bucket", func() {
	//				Expect(liveBucket.ListFilesCallCount()).To(Equal(1))
	//			})
	//
	//			By("calling copy for each file in the bucket", func() {
	//				Expect(backupBucket.CopyObjectCallCount()).To(Equal(2))
	//
	//				var keys []string
	//
	//				key, originPath, destinationPath, originBucketName, originBucketRegion := backupBucket.CopyObjectArgsForCall(1)
	//				Expect(originPath).To(Equal(""))
	//				Expect(destinationPath).To(Equal("destination-string"))
	//				Expect(originBucketName).To(Equal(configLiveBucket))
	//				Expect(originBucketRegion).To(Equal(configLiveRegion))
	//				keys = append(keys, key)
	//
	//				key, originPath, destinationPath, originBucketName, originBucketRegion = backupBucket.CopyObjectArgsForCall(0)
	//				Expect(originPath).To(Equal(""))
	//				Expect(destinationPath).To(Equal("destination-string"))
	//				Expect(originBucketName).To(Equal(configLiveBucket))
	//				Expect(originBucketRegion).To(Equal(configLiveRegion))
	//				keys = append(keys, key)
	//
	//				Expect(keys).To(ConsistOf("path1/file1", "path2/file2"))
	//			})
	//
	//			By("returning the bucketBackup of the backup bucket", func() {
	//				Expect(bucketBackup).To(Equal(unversioned.BucketSnapshot{
	//					BucketName:   configBackupBucket,
	//					BucketRegion: configBackupRegion,
	//					Path:         "destination-string",
	//					EmptyBackup:  false,
	//				}))
	//			})
	//		})
	//
	//		Context("when CopyObject fails", func() {
	//			BeforeEach(func() {
	//				backupBucket.CopyObjectReturnsOnCall(0, fmt.Errorf("cannot copy first file"))
	//				backupBucket.CopyObjectReturnsOnCall(1, fmt.Errorf("cannot copy second file"))
	//			})
	//
	//			It("should fail", func() {
	//				Expect(err).To(MatchError(ContainSubstring("cannot copy first file")))
	//				Expect(err).To(MatchError(ContainSubstring("cannot copy second file")))
	//			})
	//		})
	//	})
	//
	//	Context("When there are no files in the bucket", func() {
	//		BeforeEach(func() {
	//			liveBucket.ListFilesReturns([]string{}, nil)
	//		})
	//
	//		It("Records that information in the backup artifact", func() {
	//
	//			By("succeeding", func() {
	//				Expect(err).NotTo(HaveOccurred())
	//			})
	//
	//			By("not calling copy", func() {
	//				Expect(backupBucket.CopyObjectCallCount()).To(Equal(0))
	//			})
	//
	//			By("recording that the backup was empty", func() {
	//				Expect(snapshot).To(Equal(unversioned.BucketSnapshot{
	//					BucketName:   configBackupBucket,
	//					BucketRegion: configBackupRegion,
	//					Path:         "destination-string",
	//					EmptyBackup:  true,
	//				}))
	//			})
	//		})
	//
	//	})
	//
	//	Context("when ListFiles fails", func() {
	//		BeforeEach(func() {
	//			liveBucket.ListFilesReturns([]string{}, fmt.Errorf("cannot list files"))
	//		})
	//
	//		It("should fail", func() {
	//			Expect(err).To(MatchError("cannot list files"))
	//		})
	//	})
	//
	//})

	Describe("Restore", func() {
		JustBeforeEach(func() {
			bucketBackup = incremental.BucketBackup{
				BucketName:          "artifact_bucket_name",
				BucketRegion:        "artifact_bucket_region",
				Blobs:               []string{"my_key", "another_key"},
				BackupDirectoryPath: "2015-12-13-05-06-07/my_bucket",
			}
			err = bucketPair.Restore(bucketBackup)
		})

		BeforeEach(func() {
			blob1 := new(fakes.FakeBlob)
			blob1.PathReturns("path/to/fake/blob1")
			blob2 := new(fakes.FakeBlob)
			blob2.PathReturns("path/to/fake/blob2")
			liveBucket.ListBlobsReturns([]incremental.Blob{blob1, blob2}, nil)
		})

		It("successfully copies from the backup bucket to the live bucket", func() {
			By("not returning an error", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			By("copying from the backup location to the live location", func() {
				Expect(liveBucket.CopyBlobFromBucketCallCount()).To(Equal(2))
				// NEEDS WORK
				//var keys []string

				bucket, src, dst, region := liveBucket.CopyBlobFromBucketArgsForCall(0)
				Expect(src).To(Equal(bucketBackup.BackupDirectoryPath))
				Expect(dst).To(Equal(""))
				Expect(bucket.Name()).To(Equal(bucketBackup.BucketName))
				Expect(region).To(Equal(bucketBackup.BucketRegion))
				//keys = append(keys)

				bucket, src, dst, region = liveBucket.CopyBlobFromBucketArgsForCall(1)
				Expect(src).To(Equal(bucketBackup.BackupDirectoryPath))
				Expect(dst).To(Equal(""))
				Expect(bucket.Name()).To(Equal(bucketBackup.BucketName))
				Expect(region).To(Equal(bucketBackup.BucketRegion))
				//keys = append(keys, )

				//Expect(keys).To(ConsistOf("my_key", "another_key"))
			})
		})

		Context("When CopyObject errors", func() {
			BeforeEach(func() {
				liveBucket.CopyBlobFromBucketReturns(fmt.Errorf("cannot copy object"))
			})

			It("errors", func() {
				Expect(err).To(MatchError(ContainSubstring("cannot copy object")))
			})
		})
	})

	Describe("CheckValidity", func() {
		Context("when the live bucket and the backup bucket are not the same", func() {
			It("returns nil", func() {
				Expect(incremental.NewIncrementalBucketPair(liveBucket, backupBucket).CheckValidity()).To(BeNil())
			})
		})

		Context("when the live bucket and the backup bucket are the same", func() {
			It("returns an error", func() {
				Expect(incremental.NewIncrementalBucketPair(liveBucket, liveBucket).CheckValidity()).To(MatchError("live bucket and backup bucket cannot be the same"))
			})
		})
	})
})
