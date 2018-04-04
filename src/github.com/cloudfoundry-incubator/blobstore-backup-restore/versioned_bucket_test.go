package blobstore_test

import (
	. "github.com/cloudfoundry-incubator/blobstore-backup-restore"

	"os"

	"github.com/cloudfoundry-incubator/blobstore-backup-restore/s3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("S3VersionedBucket", func() {
	var bucketObjectUnderTest S3VersionedBucket
	var err error

	RunVersionedBucketTests := func(mainRegion, secondaryRegion, endpoint, accessKey, secretKey string) {
		var bucket TestS3Bucket

		var firstVersionOfFile1 string
		var secondVersionOfFile1 string
		var thirdVersionOfFile1 string
		var firstVersionOfFile2 string
		var deletedVersionOfFile2 string

		var creds = s3.S3AccessKey{
			Id:     accessKey,
			Secret: secretKey,
		}

		BeforeEach(func() {
			bucket = setUpVersionedS3Bucket(mainRegion, endpoint, creds)

			firstVersionOfFile1 = uploadFile(bucket.Name, endpoint, "test-1", "1-A", creds)
			secondVersionOfFile1 = uploadFile(bucket.Name, endpoint, "test-1", "1-B", creds)
			thirdVersionOfFile1 = uploadFile(bucket.Name, endpoint, "test-1", "1-C", creds)
			firstVersionOfFile2 = uploadFile(bucket.Name, endpoint, "test-2", "2-A", creds)
			deletedVersionOfFile2 = deleteFile(bucket.Name, endpoint, "test-2", creds)

			s3Bucket, err := s3.NewBucket(bucket.Name, bucket.Region, endpoint, creds)
			Expect(err).NotTo(HaveOccurred())
			bucketObjectUnderTest = NewS3VersionedBucket(s3Bucket)
		})

		AfterEach(func() {
			tearDownVersionedBucket(bucket.Name, endpoint, creds)
		})

		Describe("Versions", func() {
			var versions []s3.Version

			JustBeforeEach(func() {
				versions, err = bucketObjectUnderTest.Versions()
			})

			Context("when retrieving versions succeeds", func() {
				It("returns a list of all versions in the bucket", func() {
					Expect(err).NotTo(HaveOccurred())
					Expect(versions).To(ConsistOf(
						s3.Version{Id: firstVersionOfFile1, Key: "test-1", IsLatest: false},
						s3.Version{Id: secondVersionOfFile1, Key: "test-1", IsLatest: false},
						s3.Version{Id: thirdVersionOfFile1, Key: "test-1", IsLatest: true},
						s3.Version{Id: firstVersionOfFile2, Key: "test-2", IsLatest: false},
					))
				})
			})

			Context("when the bucket can't be reached", func() {
				BeforeEach(func() {
					s3Bucket, err := s3.NewBucket(
						bucket.Name,
						bucket.Region,
						endpoint,
						s3.S3AccessKey{Id: "NOT RIGHT", Secret: "NOT RIGHT"},
					)
					Expect(err).NotTo(HaveOccurred())
					bucketObjectUnderTest = NewS3VersionedBucket(s3Bucket)
				})

				It("returns the error", func() {
					Expect(versions).To(BeNil())
					Expect(err).To(MatchError(MatchRegexp("could not check if bucket (.*) is versioned")))
				})
			})

			Context("when the bucket is not versioned", func() {

				var unversionedBucket TestS3Bucket

				BeforeEach(func() {
					unversionedBucket = setUpS3UnversionedBucket(mainRegion, endpoint, creds)
					uploadFile(unversionedBucket.Name, endpoint, "unversioned-test", "UNVERSIONED-TEST", creds)

					s3Bucket, err := s3.NewBucket(unversionedBucket.Name, unversionedBucket.Region, endpoint, creds)
					Expect(err).NotTo(HaveOccurred())
					bucketObjectUnderTest = NewS3VersionedBucket(s3Bucket)
				})

				AfterEach(func() {
					tearDownBucket(unversionedBucket.Name, endpoint, creds)
				})

				It("fails", func() {
					Expect(versions).To(BeNil())
					Expect(err).To(MatchError(ContainSubstring("is not versioned")))
				})
			})

			Context("when the bucket has a lot of files", func() {
				BeforeEach(func() {
					s3Bucket, err := s3.NewBucket("sdk-big-bucket-integration-test", mainRegion, endpoint, creds)
					Expect(err).NotTo(HaveOccurred())
					bucketObjectUnderTest = NewS3VersionedBucket(s3Bucket)
				})

				It("works", func() {
					versions, err := bucketObjectUnderTest.Versions()

					Expect(err).NotTo(HaveOccurred())
					Expect(len(versions)).To(Equal(2001))
				})
			})
		})

		Describe("CopyVersions from same bucket", func() {
			BeforeEach(func() {
				uploadFile(bucket.Name, endpoint, "test-3", "TEST-3-A", creds)
			})

			JustBeforeEach(func() {
				err = bucketObjectUnderTest.CopyVersions(bucket.Region, bucket.Name, []BlobVersion{
					{BlobKey: "test-1", Id: secondVersionOfFile1},
					{BlobKey: "test-2", Id: firstVersionOfFile2},
				})
			})

			Context("when putting versions succeeds", func() {
				It("restores files to versions specified in the backup and does not delete pre-existing blobs", func() {
					Expect(err).NotTo(HaveOccurred())

					Expect(listFiles(bucket.Name, endpoint, creds)).To(ConsistOf(
						"test-1", "test-2", "test-3"))
					Expect(getFileContents(bucket.Name, endpoint, "test-1", creds)).To(Equal(
						"1-B"))
					Expect(getFileContents(bucket.Name, endpoint, "test-2", creds)).To(Equal(
						"2-A"))
					Expect(getFileContents(bucket.Name, endpoint, "test-3", creds)).To(Equal(
						"TEST-3-A"))
				})
			})

			Context("when putting versions fails", func() {
				BeforeEach(func() {
					s3Bucket, err := s3.NewBucket(bucket.Name, bucket.Region, endpoint, s3.S3AccessKey{})
					Expect(err).NotTo(HaveOccurred())
					bucketObjectUnderTest = NewS3VersionedBucket(s3Bucket)
				})

				It("errors", func() {
					Expect(err).To(HaveOccurred())
				})
			})

			Context("when the bucket is not versioned", func() {

				var unversionedBucket TestS3Bucket

				BeforeEach(func() {
					unversionedBucket = setUpS3UnversionedBucket(mainRegion, endpoint, creds)
					uploadFile(unversionedBucket.Name, endpoint, "unversioned-test", "UNVERSIONED-TEST", creds)

					s3Bucket, err := s3.NewBucket(unversionedBucket.Name, unversionedBucket.Region, endpoint, creds)
					Expect(err).NotTo(HaveOccurred())
					bucketObjectUnderTest = NewS3VersionedBucket(s3Bucket)
				})

				AfterEach(func() {
					tearDownBucket(unversionedBucket.Name, endpoint, creds)
				})

				It("fails", func() {
					Expect(err).To(MatchError(ContainSubstring("is not versioned")))
				})
			})
		})

		Describe("CopyVersions from different bucket in different region", func() {
			var secondaryBucket TestS3Bucket
			var versionOfFileWhichWasSubsequentlyDeleted, versionOfFileToBeRestored string

			BeforeEach(func() {
				clearOutVersionedBucket(bucket.Name, endpoint, creds)
				secondaryBucket = setUpVersionedS3Bucket(secondaryRegion, endpoint, creds)
				versionOfFileToBeRestored = uploadFile(
					secondaryBucket.Name,
					endpoint,
					"file-to-restore",
					"whatever",
					creds,
				)
				versionOfFileWhichWasSubsequentlyDeleted = uploadFile(
					secondaryBucket.Name,
					endpoint,
					"deleted-file-to-restore",
					"whatever",
					creds,
				)
				deleteFile(secondaryBucket.Name, endpoint, "deleted-file-to-restore", creds)
				uploadFile(bucket.Name, endpoint, "file-to-be-destroyed-by-restore",
					"whatever", creds)
			})

			JustBeforeEach(func() {
				err = bucketObjectUnderTest.CopyVersions(secondaryBucket.Region, secondaryBucket.Name,
					[]BlobVersion{
						{BlobKey: "file-to-restore", Id: versionOfFileToBeRestored},
						{BlobKey: "deleted-file-to-restore", Id: versionOfFileWhichWasSubsequentlyDeleted},
					})
			})

			It("restores files from the secondary to the main bucket and does not delete pre-existing blobs", func() {
				Expect(err).NotTo(HaveOccurred())

				Expect(listFiles(bucket.Name, endpoint, creds)).To(
					ConsistOf("file-to-restore", "deleted-file-to-restore", "file-to-be-destroyed-by-restore"),
				)
			})

			AfterEach(func() {
				tearDownVersionedBucket(secondaryBucket.Name, endpoint, creds)
			})
		})
	}

	Describe("AWS S3 bucket", func() {
		RunVersionedBucketTests(
			"eu-west-1",
			"us-west-1",
			"",
			os.Getenv("TEST_AWS_ACCESS_KEY_ID"),
			os.Getenv("TEST_AWS_SECRET_ACCESS_KEY"),
		)
	})

	Describe("ECS S3-compatible bucket", func() {
		RunVersionedBucketTests(
			"eu-west-1",
			"us-east-1",
			"https://object.ecstestdrive.com",
			os.Getenv("TEST_ECS_ACCESS_KEY_ID"),
			os.Getenv("TEST_ECS_SECRET_ACCESS_KEY"),
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

			creds := s3.S3AccessKey{
				Id:     os.Getenv("TEST_AWS_ACCESS_KEY_ID"),
				Secret: os.Getenv("TEST_AWS_SECRET_ACCESS_KEY"),
			}

			clearOutVersionedBucket(bucketName, endpoint, creds)
			s3Bucket, err := s3.NewBucket(bucketName, region, endpoint, creds)
			Expect(err).NotTo(HaveOccurred())
			bucketObjectUnderTest = NewS3VersionedBucket(s3Bucket)
		})

		Context("when backup an empty bucket", func() {
			It("does not fail", func() {
				_, err := bucketObjectUnderTest.Versions()
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when restore from an empty bucket", func() {
			It("does not fail", func() {
				err := bucketObjectUnderTest.CopyVersions(region, bucketName, []BlobVersion{})
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("CopyVersions with a big file on AWS", func() {
		var destinationBucket TestS3Bucket
		var region string
		var endpoint string
		var creds s3.S3AccessKey

		BeforeEach(func() {
			region = "eu-west-1"
			endpoint = ""

			creds = s3.S3AccessKey{
				Id:     os.Getenv("TEST_AWS_ACCESS_KEY_ID"),
				Secret: os.Getenv("TEST_AWS_SECRET_ACCESS_KEY"),
			}

			destinationBucket = setUpVersionedS3Bucket(region, endpoint, creds)

			s3Bucket, err := s3.NewBucket(destinationBucket.Name, region, endpoint, creds)
			Expect(err).NotTo(HaveOccurred())
			bucketObjectUnderTest = NewS3VersionedBucket(s3Bucket)
		})

		AfterEach(func() {
			clearOutVersionedBucket(destinationBucket.Name, endpoint, creds)
			tearDownBucket(destinationBucket.Name, endpoint, creds)
		})

		It("works", func() {
			err := bucketObjectUnderTest.CopyVersions(
				"eu-west-1",
				"large-blob-test-bucket", []BlobVersion{
					{BlobKey: "big_file", Id: "YfWcz5KoJzfjKB9gnBI6q7ue_jZGTvkw"},
				})

			Expect(err).NotTo(HaveOccurred())
			Expect(listFiles(destinationBucket.Name, endpoint, creds)).To(ConsistOf("big_file"))

			localFilePath := downloadFileToTmp(destinationBucket.Name, endpoint, "big_file", creds)
			Expect(shasum(localFilePath)).To(Equal("188f500de28479d67e7375566750472e58e4cec1"))
		})
	})
})
