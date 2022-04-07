package s3bucket_test

import (
	"github.com/cloudfoundry/backup-and-restore-sdk-release/src/s3-blobstore-backup-restore/s3bucket"
	"github.com/cloudfoundry/backup-and-restore-sdk-release/src/s3-blobstore-backup-restore/versioned"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("VersionedBucket", func() {
	var bucketObjectUnderTest versioned.Bucket
	var err error

	Describe("S3 bucket", func() {
		var bucketName string

		var firstVersionOfFile1 string
		var secondVersionOfFile1 string
		var thirdVersionOfFile1 string
		var firstVersionOfFile2 string

		var creds = s3bucket.AccessKey{
			Id:     AccessKey,
			Secret: SecretKey,
		}

		BeforeEach(func() {
			bucketName = setUpVersionedBucket(LiveRegion, S3Endpoint, creds)

			firstVersionOfFile1 = uploadFile(bucketName, S3Endpoint, "test-1", "1-A", creds)
			secondVersionOfFile1 = uploadFile(bucketName, S3Endpoint, "test-1", "1-B", creds)
			thirdVersionOfFile1 = uploadFile(bucketName, S3Endpoint, "test-1", "1-C", creds)
			firstVersionOfFile2 = uploadFile(bucketName, S3Endpoint, "test-2", "2-A", creds)
			deleteFile(bucketName, S3Endpoint, "test-2", creds)

			bucketObjectUnderTest, err = s3bucket.NewBucket(bucketName, LiveRegion, S3Endpoint, creds, false, false)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			tearDownVersionedBucket(bucketName, S3Endpoint, creds)
		})

		Describe("ListVersions", func() {
			var versions []s3bucket.Version

			JustBeforeEach(func() {
				versions, err = bucketObjectUnderTest.ListVersions()
			})

			Context("when retrieving versions succeeds", func() {
				It("returns a list of all versions in the bucket", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(versions).To(ConsistOf(
						s3bucket.Version{Id: firstVersionOfFile1, Key: "test-1", IsLatest: false},
						s3bucket.Version{Id: secondVersionOfFile1, Key: "test-1", IsLatest: false},
						s3bucket.Version{Id: thirdVersionOfFile1, Key: "test-1", IsLatest: true},
						s3bucket.Version{Id: firstVersionOfFile2, Key: "test-2", IsLatest: false},
					))
				})
			})

			Context("when the bucket can't be reached", func() {
				BeforeEach(func() {
					bucketObjectUnderTest, err = s3bucket.NewBucket(
						bucketName,
						LiveRegion,
						S3Endpoint,
						s3bucket.AccessKey{Id: "NOT RIGHT", Secret: "NOT RIGHT"},
						false,
						true,
					)
					Expect(err).NotTo(HaveOccurred())
				})

				It("returns the error", func() {
					Expect(versions).To(BeNil())
					Expect(err).To(MatchError(MatchRegexp("could not check if bucket (.*) is versioned")))
				})
			})

			Context("when the bucket is not versioned", func() {
				var unversionedBucketName string

				BeforeEach(func() {
					unversionedBucketName = setUpUnversionedBucket(LiveRegion, S3Endpoint, creds)
					uploadFile(unversionedBucketName, S3Endpoint, "unversioned-test", "UNVERSIONED-TEST", creds)

					bucketObjectUnderTest, err = s3bucket.NewBucket(unversionedBucketName, LiveRegion, S3Endpoint, creds, false, false)
					Expect(err).NotTo(HaveOccurred())
				})

				AfterEach(func() {
					tearDownBucket(unversionedBucketName, S3Endpoint, creds)
				})

				It("fails", func() {
					Expect(err).To(MatchError(ContainSubstring("is not versioned")))
				})
			})

			Context("when the bucket has a lot of files", func() {
				BeforeEach(func() {
					bucketObjectUnderTest, err = s3bucket.NewBucket("sdk-big-bucket-integration-test", LiveRegion, S3Endpoint, creds, false, false)
					Expect(err).NotTo(HaveOccurred())
				})

				It("works", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(len(versions)).To(Equal(2001))
				})
			})

			Context("when the bucket is empty", func() {
				BeforeEach(func() {
					clearOutVersionedBucket(bucketName, S3Endpoint, creds)
				})

				It("works", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(len(versions)).To(Equal(0))
				})
			})
		})

		Describe("IsVersioned", func() {
			var (
				unversionedBucketName string
				bucketObjectUnderTest versioned.Bucket
			)
			Context("when the bucket is versioned", func() {

				BeforeEach(func() {
					unversionedBucketName = setUpVersionedBucket(LiveRegion, S3Endpoint, creds)
					uploadFile(unversionedBucketName, S3Endpoint, "unversioned-test", "UNVERSIONED-TEST", creds)

					bucketObjectUnderTest, err = s3bucket.NewBucket(unversionedBucketName, LiveRegion, S3Endpoint, creds, false, false)
					Expect(err).NotTo(HaveOccurred())
				})

				AfterEach(func() {
					tearDownVersionedBucket(unversionedBucketName, S3Endpoint, creds)
				})

				It("returns true", func() {
					isVersioned, err := bucketObjectUnderTest.IsVersioned()
					Expect(err).NotTo(HaveOccurred())
					Expect(isVersioned).To(BeTrue())
				})
			})

			Context("when the bucket is not versioned", func() {
				BeforeEach(func() {
					unversionedBucketName = setUpUnversionedBucket(LiveRegion, S3Endpoint, creds)
					uploadFile(unversionedBucketName, S3Endpoint, "unversioned-test", "UNVERSIONED-TEST", creds)

					bucketObjectUnderTest, err = s3bucket.NewBucket(unversionedBucketName, LiveRegion, S3Endpoint, creds, false, false)
					Expect(err).NotTo(HaveOccurred())
				})

				AfterEach(func() {
					tearDownBucket(unversionedBucketName, S3Endpoint, creds)
				})

				It("returns false", func() {
					isVersioned, err := bucketObjectUnderTest.IsVersioned()
					Expect(err).NotTo(HaveOccurred())
					Expect(isVersioned).To(BeFalse())
				})
			})

			Context("when it fails to check the version", func() {
				BeforeEach(func() {
					bucketObjectUnderTest, err = s3bucket.NewBucket("does-not-exist", LiveRegion, S3Endpoint, creds, false, false)
					Expect(err).NotTo(HaveOccurred())
				})

				It("returns an error", func() {
					_, err := bucketObjectUnderTest.IsVersioned()
					Expect(err).To(MatchError(ContainSubstring("could not check if bucket does-not-exist is versioned")))
				})
			})
		})

		Describe("CopyVersion from same bucket", func() {
			JustBeforeEach(func() {
				err = bucketObjectUnderTest.CopyVersion("test-1", secondVersionOfFile1, bucketName, LiveRegion)
			})

			Context("when putting versions succeeds", func() {
				It("restores files to versions specified in the backup and does not delete pre-existing blobs", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(getFileContents(bucketName, S3Endpoint, "test-1", creds)).To(Equal("1-B"))
				})
			})

			Context("when putting versions fails", func() {
				BeforeEach(func() {
					bucketObjectUnderTest, err = s3bucket.NewBucket(bucketName, LiveRegion, S3Endpoint, s3bucket.AccessKey{}, false, false)
					Expect(err).NotTo(HaveOccurred())
				})

				It("errors", func() {
					Expect(err).To(HaveOccurred())
				})
			})
		})

		Describe("CopyVersion from different bucket in different region", func() {
			var secondaryBucketName string
			var versionOfFileWhichWasSubsequentlyDeleted string

			BeforeEach(func() {
				clearOutVersionedBucket(bucketName, S3Endpoint, creds)
				secondaryBucketName = setUpVersionedBucket(BackupRegion, S3Endpoint, creds)
				versionOfFileWhichWasSubsequentlyDeleted = uploadFile(
					secondaryBucketName,
					S3Endpoint,
					"deleted-file-to-restore",
					"file-contents",
					creds,
				)
				deleteFile(secondaryBucketName, S3Endpoint, "deleted-file-to-restore", creds)
			})

			JustBeforeEach(func() {
				err = bucketObjectUnderTest.CopyVersion(
					"deleted-file-to-restore",
					versionOfFileWhichWasSubsequentlyDeleted,
					secondaryBucketName,
					BackupRegion,
				)
			})

			It("restores files from the secondary to the main bucket and does not delete pre-existing blobs", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(getFileContents(bucketName, S3Endpoint, "deleted-file-to-restore", creds)).To(Equal("file-contents"))
			})

			AfterEach(func() {
				tearDownVersionedBucket(secondaryBucketName, S3Endpoint, creds)
			})
		})
	})

	Describe("CopyVersion with a big file on AWS", func() {
		var destinationBucketName string
		var creds s3bucket.AccessKey

		BeforeEach(func() {
			creds = s3bucket.AccessKey{
				Id:     AccessKey,
				Secret: SecretKey,
			}

			destinationBucketName = setUpVersionedBucket(LiveRegion, S3Endpoint, creds)

			bucketObjectUnderTest, err = s3bucket.NewBucket(destinationBucketName, LiveRegion, S3Endpoint, creds, false, false)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			clearOutVersionedBucket(destinationBucketName, S3Endpoint, creds)
			tearDownBucket(destinationBucketName, S3Endpoint, creds)
		})

		It("works", func() {
			err := bucketObjectUnderTest.CopyVersion(
				"big_file",
				"YfWcz5KoJzfjKB9gnBI6q7ue_jZGTvkw",
				"large-blob-test-bucket",
				"eu-west-1",
			)

			Expect(err).NotTo(HaveOccurred())
			Expect(listFiles(destinationBucketName, S3Endpoint, creds)).To(ConsistOf("big_file"))

			localFilePath := downloadFileToTmp(destinationBucketName, S3Endpoint, "big_file", creds)
			Expect(shasum(localFilePath)).To(Equal("188f500de28479d67e7375566750472e58e4cec1"))
		})
	})
})
