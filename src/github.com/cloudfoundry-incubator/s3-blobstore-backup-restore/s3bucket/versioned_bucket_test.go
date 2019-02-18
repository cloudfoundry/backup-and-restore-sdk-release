package s3bucket_test

import (
	"github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/s3bucket"
	"github.com/cloudfoundry-incubator/s3-blobstore-backup-restore/versioned"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("VersionedBucket", func() {
	var bucketObjectUnderTest versioned.Bucket
	var err error

	RunVersionedBucketTests := func(mainRegion, secondaryRegion, endpoint, accessKey, secretKey string) {
		var bucketName string

		var firstVersionOfFile1 string
		var secondVersionOfFile1 string
		var thirdVersionOfFile1 string
		var firstVersionOfFile2 string

		var creds = s3bucket.AccessKey{
			Id:     accessKey,
			Secret: secretKey,
		}

		BeforeEach(func() {
			bucketName = setUpVersionedBucket(mainRegion, endpoint, creds)

			firstVersionOfFile1 = uploadFile(bucketName, endpoint, "test-1", "1-A", creds)
			secondVersionOfFile1 = uploadFile(bucketName, endpoint, "test-1", "1-B", creds)
			thirdVersionOfFile1 = uploadFile(bucketName, endpoint, "test-1", "1-C", creds)
			firstVersionOfFile2 = uploadFile(bucketName, endpoint, "test-2", "2-A", creds)
			deleteFile(bucketName, endpoint, "test-2", creds)

			bucketObjectUnderTest, err = s3bucket.NewBucket(bucketName, mainRegion, endpoint, creds, false)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			tearDownVersionedBucket(bucketName, endpoint, creds)
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
						mainRegion,
						endpoint,
						s3bucket.AccessKey{Id: "NOT RIGHT", Secret: "NOT RIGHT"},
						false,
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
					unversionedBucketName = setUpUnversionedBucket(mainRegion, endpoint, creds)
					uploadFile(unversionedBucketName, endpoint, "unversioned-test", "UNVERSIONED-TEST", creds)

					bucketObjectUnderTest, err = s3bucket.NewBucket(unversionedBucketName, mainRegion, endpoint, creds, false)
					Expect(err).NotTo(HaveOccurred())
				})

				AfterEach(func() {
					tearDownBucket(unversionedBucketName, endpoint, creds)
				})

				It("fails", func() {
					Expect(versions).To(BeNil())
					Expect(err).To(MatchError(ContainSubstring("is not versioned")))
				})
			})

			Context("when the bucket has a lot of files", func() {
				BeforeEach(func() {
					bucketObjectUnderTest, err = s3bucket.NewBucket("sdk-big-bucket-integration-test", mainRegion, endpoint, creds, false)
					Expect(err).NotTo(HaveOccurred())
				})

				It("works", func() {
					versions, err := bucketObjectUnderTest.ListVersions()

					Expect(err).NotTo(HaveOccurred())
					Expect(len(versions)).To(Equal(2001))
				})
			})
		})

		Describe("CheckIfVersioned", func() {
			Context("when the bucket is not versioned", func() {
				var unversionedBucketName string
				var bucketObjectUnderTest versioned.Bucket

				BeforeEach(func() {
					unversionedBucketName = setUpUnversionedBucket(mainRegion, endpoint, creds)
					uploadFile(unversionedBucketName, endpoint, "unversioned-test", "UNVERSIONED-TEST", creds)

					bucketObjectUnderTest, err = s3bucket.NewBucket(unversionedBucketName, mainRegion, endpoint, creds, false)
					Expect(err).NotTo(HaveOccurred())
				})

				AfterEach(func() {
					tearDownBucket(unversionedBucketName, endpoint, creds)
				})

				It("fails", func() {
					err = bucketObjectUnderTest.CheckIfVersioned()
					Expect(err).To(MatchError(ContainSubstring("is not versioned")))
				})
			})
		})

		Describe("CopyVersion from same bucket", func() {
			JustBeforeEach(func() {
				err = bucketObjectUnderTest.CopyVersion("test-1", secondVersionOfFile1, bucketName, mainRegion)
			})

			Context("when putting versions succeeds", func() {
				It("restores files to versions specified in the backup and does not delete pre-existing blobs", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(getFileContents(bucketName, endpoint, "test-1", creds)).To(Equal("1-B"))
				})
			})

			Context("when putting versions fails", func() {
				BeforeEach(func() {
					bucketObjectUnderTest, err = s3bucket.NewBucket(bucketName, mainRegion, endpoint, s3bucket.AccessKey{}, false)
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
				clearOutVersionedBucket(bucketName, endpoint, creds)
				secondaryBucketName = setUpVersionedBucket(secondaryRegion, endpoint, creds)
				versionOfFileWhichWasSubsequentlyDeleted = uploadFile(
					secondaryBucketName,
					endpoint,
					"deleted-file-to-restore",
					"file-contents",
					creds,
				)
				deleteFile(secondaryBucketName, endpoint, "deleted-file-to-restore", creds)
			})

			JustBeforeEach(func() {
				err = bucketObjectUnderTest.CopyVersion(
					"deleted-file-to-restore",
					versionOfFileWhichWasSubsequentlyDeleted,
					secondaryBucketName,
					secondaryRegion,
				)
			})

			It("restores files from the secondary to the main bucket and does not delete pre-existing blobs", func() {
				Expect(err).NotTo(HaveOccurred())
				Expect(getFileContents(bucketName, endpoint, "deleted-file-to-restore", creds)).To(Equal("file-contents"))
			})

			AfterEach(func() {
				tearDownVersionedBucket(secondaryBucketName, endpoint, creds)
			})
		})
	}

	Describe("AWS S3 bucket", func() {
		RunVersionedBucketTests(
			"eu-west-1",
			"us-west-1",
			"",
			TestAWSAccessKeyID,
			TestAWSSecretAccessKey,
		)
	})

	Describe("ECS S3-compatible bucket", func() {
		RunVersionedBucketTests(
			"eu-west-1",
			"us-east-1",
			"https://object.ecstestdrive.com",
			TestECSAccessKeyID,
			TestECSSecretAccessKey,
		)
	})

	Describe("Empty AWS S3 bucket", func() {
		var region string
		var bucketName string
		var endpoint string

		BeforeEach(func() {
			region = "eu-west-1"
			bucketName = "bbr-integration-test-bucket-empty"
			endpoint = ""

			creds := s3bucket.AccessKey{
				Id:     TestAWSAccessKeyID,
				Secret: TestAWSSecretAccessKey,
			}

			clearOutVersionedBucket(bucketName, endpoint, creds)
			bucketObjectUnderTest, err = s3bucket.NewBucket(bucketName, region, endpoint, creds, false)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when listing versions for an empty bucket", func() {
			It("does not fail", func() {
				_, err := bucketObjectUnderTest.ListVersions()
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("CopyVersion with a big file on AWS", func() {
		var destinationBucketName string
		var region string
		var endpoint string
		var creds s3bucket.AccessKey

		BeforeEach(func() {
			region = "eu-west-1"
			endpoint = ""

			creds = s3bucket.AccessKey{
				Id:     TestAWSAccessKeyID,
				Secret: TestAWSSecretAccessKey,
			}

			destinationBucketName = setUpVersionedBucket(region, endpoint, creds)

			bucketObjectUnderTest, err = s3bucket.NewBucket(destinationBucketName, region, endpoint, creds, false)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			clearOutVersionedBucket(destinationBucketName, endpoint, creds)
			tearDownBucket(destinationBucketName, endpoint, creds)
		})

		It("works", func() {
			err := bucketObjectUnderTest.CopyVersion(
				"big_file",
				"YfWcz5KoJzfjKB9gnBI6q7ue_jZGTvkw",
				"large-blob-test-bucket",
				"eu-west-1",
			)

			Expect(err).NotTo(HaveOccurred())
			Expect(listFiles(destinationBucketName, endpoint, creds)).To(ConsistOf("big_file"))

			localFilePath := downloadFileToTmp(destinationBucketName, endpoint, "big_file", creds)
			Expect(shasum(localFilePath)).To(Equal("188f500de28479d67e7375566750472e58e4cec1"))
		})
	})
})
