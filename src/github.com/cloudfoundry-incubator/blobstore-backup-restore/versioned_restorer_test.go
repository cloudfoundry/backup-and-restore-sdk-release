package blobstore_test

import (
	. "github.com/cloudfoundry-incubator/blobstore-backup-restore"

	"errors"

	"github.com/cloudfoundry-incubator/blobstore-backup-restore/fakes"
	s3fakes "github.com/cloudfoundry-incubator/blobstore-backup-restore/s3/fakes"

	"github.com/cloudfoundry-incubator/blobstore-backup-restore/s3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("VersionedRestorer", func() {
	var dropletsBucket *s3fakes.FakeVersionedBucket
	var buildpacksBucket *s3fakes.FakeVersionedBucket
	var packagesBucket *s3fakes.FakeVersionedBucket

	var artifact *fakes.FakeVersionedArtifact

	var err error

	var restorer VersionedRestorer

	BeforeEach(func() {
		dropletsBucket = new(s3fakes.FakeVersionedBucket)
		buildpacksBucket = new(s3fakes.FakeVersionedBucket)
		packagesBucket = new(s3fakes.FakeVersionedBucket)

		artifact = new(fakes.FakeVersionedArtifact)

		restorer = NewVersionedRestorer(map[string]s3.VersionedBucket{
			"droplets":   dropletsBucket,
			"buildpacks": buildpacksBucket,
			"packages":   packagesBucket,
		}, artifact)
	})

	JustBeforeEach(func() {
		err = restorer.Run()
	})

	Context("when the artifact is valid and copying versions to buckets works", func() {
		BeforeEach(func() {
			artifact.LoadReturns(map[string]BucketSnapshot{
				"droplets": {
					BucketName: "my_droplets_bucket",
					RegionName: "my_droplets_region",
					Versions: []BlobVersion{
						{BlobKey: "one", Id: "13"},
						{BlobKey: "two", Id: "22"},
					},
				},
				"buildpacks": {
					BucketName: "my_buildpacks_bucket",
					RegionName: "my_buildpacks_region",
					Versions: []BlobVersion{
						{BlobKey: "three", Id: "32"},
					},
				},
				"packages": {
					BucketName: "my_packages_bucket",
					RegionName: "my_packages_region",
					Versions: []BlobVersion{
						{BlobKey: "four", Id: "43"},
					},
				},
			}, nil)

			dropletsBucket.CopyVersionReturns(nil)
			buildpacksBucket.CopyVersionReturns(nil)
			packagesBucket.CopyVersionReturns(nil)
		})

		It("restores a backup to the corresponding buckets", func() {
			By("successfully running", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			By("Checking the buckets are versioned", func() {
				Expect(dropletsBucket.CheckIfVersionedCallCount()).To(Equal(1))
				Expect(buildpacksBucket.CheckIfVersionedCallCount()).To(Equal(1))
				Expect(packagesBucket.CheckIfVersionedCallCount()).To(Equal(1))
			})

			By("Calling CopyVersion for each object in the droplets bucket", func() {
				Expect(dropletsBucket.CopyVersionCallCount()).To(Equal(2))

				expectedBlobKey, expectedVersionId, expectedSourceBucketName, expectedSourceRegionName := dropletsBucket.CopyVersionArgsForCall(0)
				Expect(expectedBlobKey).To(Equal("one"))
				Expect(expectedVersionId).To(Equal("13"))
				Expect(expectedSourceBucketName).To(Equal("my_droplets_bucket"))
				Expect(expectedSourceRegionName).To(Equal("my_droplets_region"))

				expectedBlobKey, expectedVersionId, expectedSourceBucketName, expectedSourceRegionName = dropletsBucket.CopyVersionArgsForCall(1)
				Expect(expectedBlobKey).To(Equal("two"))
				Expect(expectedVersionId).To(Equal("22"))
				Expect(expectedSourceBucketName).To(Equal("my_droplets_bucket"))
				Expect(expectedSourceRegionName).To(Equal("my_droplets_region"))
			})

			By("Calling CopyVersions for each object in the buildpacks bucket", func() {
				Expect(buildpacksBucket.CopyVersionCallCount()).To(Equal(1))

				expectedBlobKey, expectedVersionId, expectedSourceBucketName, expectedSourceRegionName := buildpacksBucket.CopyVersionArgsForCall(0)
				Expect(expectedBlobKey).To(Equal("three"))
				Expect(expectedVersionId).To(Equal("32"))
				Expect(expectedSourceBucketName).To(Equal("my_buildpacks_bucket"))
				Expect(expectedSourceRegionName).To(Equal("my_buildpacks_region"))
			})

			By("Calling CopyVersions for each object in the packages bucket", func() {
				Expect(packagesBucket.CopyVersionCallCount()).To(Equal(1))

				expectedBlobKey, expectedVersionId, expectedSourceBucketName, expectedSourceRegionName := packagesBucket.CopyVersionArgsForCall(0)
				Expect(expectedBlobKey).To(Equal("four"))
				Expect(expectedVersionId).To(Equal("43"))
				Expect(expectedSourceBucketName).To(Equal("my_packages_bucket"))
				Expect(expectedSourceRegionName).To(Equal("my_packages_region"))
			})
		})
	})

	Context("when the artifact fails to load", func() {
		BeforeEach(func() {
			artifact.LoadReturns(nil, errors.New("artifact failed to load"))
		})

		It("stops and returns an error", func() {
			Expect(err).To(MatchError("artifact failed to load"))
			Expect(dropletsBucket.CopyVersionCallCount()).To(Equal(0))
			Expect(buildpacksBucket.CopyVersionCallCount()).To(Equal(0))
			Expect(packagesBucket.CopyVersionCallCount()).To(Equal(0))
		})
	})

	Context("when copying versions on a bucket fails", func() {
		BeforeEach(func() {
			artifact.LoadReturns(map[string]BucketSnapshot{
				"droplets": {
					BucketName: "my_droplets_bucket",
					RegionName: "my_droplets_region",
					Versions: []BlobVersion{
						{BlobKey: "one", Id: "13"},
						{BlobKey: "two", Id: "22"},
					},
				},
				"buildpacks": {
					BucketName: "my_buildpacks_bucket",
					RegionName: "my_buildpacks_region",
					Versions: []BlobVersion{
						{BlobKey: "three", Id: "32"},
					},
				},
				"packages": {
					BucketName: "my_packages_bucket",
					RegionName: "my_packages_region",
					Versions: []BlobVersion{
						{BlobKey: "four", Id: "43"},
					},
				},
			}, nil)

			dropletsBucket.CopyVersionReturns(nil)
			buildpacksBucket.CopyVersionReturns(errors.New("failed to put version to bucket 'buildpacks'"))
			packagesBucket.CopyVersionReturns(nil)
		})

		It("stops and returns an error", func() {
			Expect(err).To(MatchError("failed to put version to bucket 'buildpacks'"))
		})
	})

	Context("when there isn't a corresponding bucket recorded in backup artifact", func() {
		BeforeEach(func() {
			artifact.LoadReturns(map[string]BucketSnapshot{
				"droplets": {
					BucketName: "my_droplets_bucket",
					RegionName: "my_droplets_region",
					Versions: []BlobVersion{
						{BlobKey: "one", Id: "13"},
						{BlobKey: "two", Id: "22"},
					},
				},
				"buildpacks": {
					BucketName: "my_buildpacks_bucket",
					RegionName: "my_buildpacks_region",
					Versions: []BlobVersion{
						{BlobKey: "three", Id: "32"},
					},
				},
			}, nil)

			restorer = NewVersionedRestorer(map[string]s3.VersionedBucket{
				"droplets":   dropletsBucket,
				"buildpacks": buildpacksBucket,
				"packages":   packagesBucket,
			}, artifact)
		})

		It("fails and returns a useful error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("no entry found in backup artifact for bucket: packages"))
		})
	})
})
