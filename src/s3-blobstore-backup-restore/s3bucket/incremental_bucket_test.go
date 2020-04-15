package s3bucket_test

import (
	"fmt"

	"s3-blobstore-backup-restore/s3bucket"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("IncrementalBucket", func() {

	Describe("S3 buckets", func() {
		var (
			liveBucketName string
			liveBucket     s3bucket.Bucket
			creds          s3bucket.AccessKey
		)

		BeforeEach(func() {

			creds = s3bucket.AccessKey{Id: AccessKey, Secret: SecretKey}
			liveBucketName = setUpUnversionedBucket(LiveRegion, S3Endpoint, creds)
			uploadFile(liveBucketName, S3Endpoint, "path1/blob1", "blob1-content", creds)
			uploadFile(liveBucketName, S3Endpoint, "live/location/leaf/node", "", creds)
			uploadFile(liveBucketName, S3Endpoint, "path2/blob2", "", creds)

			var err error
			liveBucket, err = s3bucket.NewBucket(liveBucketName, LiveRegion, S3Endpoint, creds, false, s3bucket.ForcePathStyleDuringTheRefactor)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			tearDownBucket(liveBucketName, S3Endpoint, creds)
		})

		Describe("ListBlobs", func() {
			Context("without a prefix", func() {
				It("lists all the blobs in the bucket", func() {
					blobs, err := liveBucket.ListBlobs("")

					blob1 := s3bucket.NewBlob("path1/blob1")
					blob2 := s3bucket.NewBlob("live/location/leaf/node")
					blob3 := s3bucket.NewBlob("path2/blob2")

					Expect(err).NotTo(HaveOccurred())
					Expect(blobs).To(ConsistOf(blob1, blob2, blob3))
				})

				Context("when s3 list-objects errors", func() {
					It("errors", func() {
						bucketObjectUnderTest, err := s3bucket.NewBucket("does-not-exist", LiveRegion, S3Endpoint, creds, false, s3bucket.ForcePathStyleDuringTheRefactor)
						Expect(err).NotTo(HaveOccurred())

						_, err = bucketObjectUnderTest.ListBlobs("")

						Expect(err).To(MatchError(ContainSubstring("failed to list blobs from bucket does-not-exist")))
					})
				})

				Context("when the bucket has a lot of blobs", func() {
					It("works", func() {
						bucketObjectUnderTest, err := s3bucket.NewBucket("sdk-unversioned-big-bucket-integration-test", LiveRegion, S3Endpoint, creds, false, s3bucket.ForcePathStyleDuringTheRefactor)
						Expect(err).NotTo(HaveOccurred())

						blobs, err := bucketObjectUnderTest.ListBlobs("")

						Expect(err).NotTo(HaveOccurred())
						Expect(len(blobs)).To(Equal(2001))
					})
				})
			})

			Context("with a prefix", func() {
				It("lists all the blobs in the directory", func() {
					blobs, err := liveBucket.ListBlobs("live/location")

					Expect(err).NotTo(HaveOccurred())
					Expect(blobs).To(ConsistOf(s3bucket.NewBlob("leaf/node")))
				})
			})
		})

		Describe("ListDirectories", func() {
			Context("when there are several directories", func() {
				BeforeEach(func() {
					uploadFile(liveBucketName, S3Endpoint, "path1/another-blob", "", creds)
					uploadFile(liveBucketName, S3Endpoint, "top-level-blob", "", creds)
				})

				It("lists the top-level directories", func() {
					dirs, err := liveBucket.ListDirectories()

					Expect(err).NotTo(HaveOccurred())
					Expect(dirs).To(ConsistOf("path1", "live", "path2"))
				})
			})

			Context("when s3 list-objects errors", func() {
				It("errors", func() {
					bucketObjectUnderTest, err := s3bucket.NewBucket("does-not-exist", LiveRegion, S3Endpoint, creds, false, s3bucket.ForcePathStyleDuringTheRefactor)
					Expect(err).NotTo(HaveOccurred())

					_, err = bucketObjectUnderTest.ListDirectories()
					Expect(err).To(MatchError(ContainSubstring("failed to list directories from bucket does-not-exist")))
				})
			})
		})

		Describe("CopyBlobWithinBucket", func() {
			It("copies the blob", func() {
				blobs := listFiles(liveBucketName, S3Endpoint, creds)
				copyPath := "path1/blob1-copy"
				Expect(blobs).NotTo(ContainElement(copyPath))

				err := liveBucket.CopyBlobWithinBucket("path1/blob1", copyPath)

				Expect(err).NotTo(HaveOccurred())
				actualContents := getFileContents(liveBucketName, S3Endpoint, copyPath, creds)
				Expect(actualContents).To(Equal("blob1-content"))
			})

			Context("when the blob does not exist", func() {
				It("errors", func() {
					err := liveBucket.CopyBlobWithinBucket("does-not-exist", "copy-of-does-not-exist")

					Expect(err).To(MatchError(
						ContainSubstring(fmt.Sprintf("failed to get blob size for blob 'does-not-exist' in bucket '%s'", liveBucketName)),
					))
				})
			})

			Context("when the bucket was previously versioned but now is unversioned", func() {
				var (
					previouslyVersionedLiveBucketName string
					previouslyVersionedLiveBucket     s3bucket.Bucket
				)

				BeforeEach(func() {
					previouslyVersionedLiveBucketName = setUpVersionedBucket(LiveRegion, S3Endpoint, creds)
					uploadFile(previouslyVersionedLiveBucketName, S3Endpoint, "path1/blob1", "blob1-content", creds)

					disableBucketVersioning(previouslyVersionedLiveBucketName, S3Endpoint, creds)

					var err error
					previouslyVersionedLiveBucket, err = s3bucket.NewBucket(previouslyVersionedLiveBucketName, LiveRegion, S3Endpoint, creds, false, s3bucket.ForcePathStyleDuringTheRefactor)
					Expect(err).NotTo(HaveOccurred())
				})

				AfterEach(func() {
					tearDownVersionedBucket(previouslyVersionedLiveBucketName, S3Endpoint, creds)
				})

				It("successfully copies the blob", func() {
					blobs := listFiles(previouslyVersionedLiveBucketName, S3Endpoint, creds)
					copyPath := "path1/blob1-copy"
					Expect(blobs).NotTo(ContainElement(copyPath))

					err := previouslyVersionedLiveBucket.CopyBlobWithinBucket("path1/blob1", copyPath)

					Expect(err).NotTo(HaveOccurred())
					actualContents := getFileContents(previouslyVersionedLiveBucketName, S3Endpoint, copyPath, creds)
					Expect(actualContents).To(Equal("blob1-content"))
				})
			})
		})

		Describe("CopyBlobFromBucket", func() {
			var (
				backupBucketName string
				backupBucket     s3bucket.Bucket
			)

			BeforeEach(func() {
				creds = s3bucket.AccessKey{Id: AccessKey, Secret: SecretKey}
				backupBucketName = setUpUnversionedBucket(BackupRegion, S3Endpoint, creds)
				var err error
				backupBucket, err = s3bucket.NewBucket(backupBucketName, BackupRegion, S3Endpoint, creds, false, s3bucket.ForcePathStyleDuringTheRefactor)

				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				tearDownBucket(backupBucketName, S3Endpoint, creds)
			})

			It("copies the blob", func() {
				blobPath := "path1/blob1"

				err := backupBucket.CopyBlobFromBucket(liveBucket, blobPath, blobPath)

				Expect(err).NotTo(HaveOccurred())
				actualContents := getFileContents(backupBucketName, S3Endpoint, blobPath, creds)
				Expect(actualContents).To(Equal("blob1-content"))
			})

			Context("when the blob does not exist", func() {
				It("errors", func() {
					err := backupBucket.CopyBlobFromBucket(liveBucket, "does-not-exist", "copy-of-does-not-exist")

					Expect(err).To(MatchError(
						ContainSubstring(fmt.Sprintf("failed to get blob size for blob 'does-not-exist' in bucket '%s'", liveBucketName)),
					))
				})
			})
		})

		Describe("UploadBlob", func() {
			It("uploads the blob", func() {
				blobPath := "some/blob"

				err := liveBucket.UploadBlob(blobPath, "blob contents")

				Expect(err).NotTo(HaveOccurred())
				actualContents := getFileContents(liveBucketName, S3Endpoint, blobPath, creds)
				Expect(actualContents).To(Equal("blob contents"))
			})

			Context("when the bucket does not exist", func() {
				It("errors", func() {
					bucket, err := s3bucket.NewBucket("does-not-exist", LiveRegion, S3Endpoint, creds, false, s3bucket.ForcePathStyleDuringTheRefactor)
					Expect(err).NotTo(HaveOccurred())

					err = bucket.UploadBlob("some/blob", "blob contents")

					Expect(err).To(MatchError(ContainSubstring("failed to upload blob")))
				})
			})
		})

		Describe("HasBlob", func() {
			Context("when the blob exists", func() {
				It("returns true", func() {
					exists, err := liveBucket.HasBlob("path1/blob1")
					Expect(err).NotTo(HaveOccurred())

					Expect(exists).To(BeTrue())
				})
			})

			Context("when the blob does not exist", func() {
				It("returns false", func() {
					exists, err := liveBucket.HasBlob("does-not-exist-blob")
					Expect(err).NotTo(HaveOccurred())

					Expect(exists).To(BeFalse())
				})
			})

			Context("when the bucket does not exist", func() {
				It("errors", func() {
					bucket, err := s3bucket.NewBucket("does-not-exist", LiveRegion, S3Endpoint, creds, false, s3bucket.ForcePathStyleDuringTheRefactor)
					Expect(err).NotTo(HaveOccurred())
					_, err = bucket.HasBlob("does-not-exist-blob")

					Expect(err).To(MatchError(ContainSubstring("failed to check if blob exists")))
				})
			})
		})
	})

	Describe("S3 large file test", func() {
		var (
			creds      s3bucket.AccessKey
			bucketName string
		)

		BeforeEach(func() {
			creds = s3bucket.AccessKey{
				Id:     AccessKey,
				Secret: SecretKey,
			}

			bucketName = setUpUnversionedBucket("eu-west-1", S3Endpoint, creds)
		})

		AfterEach(func() {
			tearDownBucket(bucketName, S3Endpoint, creds)
		})

		It("works", func() {
			bucket, err := s3bucket.NewBucket(bucketName, LiveRegion, S3Endpoint, creds, false, s3bucket.ForcePathStyleDuringTheRefactor)
			Expect(err).NotTo(HaveOccurred())

			srcBucket, err := s3bucket.NewBucket(PreExistingBigFileBucketName, LiveRegion, S3Endpoint, creds, false, s3bucket.ForcePathStyleDuringTheRefactor)
			Expect(err).NotTo(HaveOccurred())

			err = bucket.CopyBlobFromBucket(srcBucket, "big_file", "path/to/big_file")

			By("succeeding")
			Expect(err).NotTo(HaveOccurred())

			By("copying the large file")
			Expect(listFiles(bucketName, S3Endpoint, creds)).To(ConsistOf("path/to/big_file"))

			By("not corrupting the large file")
			Expect(
				shasum(downloadFileToTmp(bucketName, S3Endpoint, "path/to/big_file", creds))).To(
				Equal("91d50642dd930e9542c39d36f0516d45f4e1af0d"))
		})
	})
})
