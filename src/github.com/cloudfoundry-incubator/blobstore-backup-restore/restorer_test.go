package blobstore_test

import (
	. "github.com/cloudfoundry-incubator/blobstore-backup-restore"

	"errors"

	"github.com/cloudfoundry-incubator/blobstore-backup-restore/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("VersionedRestorer", func() {
	var dropletsBucket *fakes.FakeVersionedBucket
	var buildpacksBucket *fakes.FakeVersionedBucket
	var packagesBucket *fakes.FakeVersionedBucket

	var artifact *fakes.FakeVersionedArtifact

	var err error

	var restorer VersionedRestorer

	BeforeEach(func() {
		dropletsBucket = new(fakes.FakeVersionedBucket)
		buildpacksBucket = new(fakes.FakeVersionedBucket)
		packagesBucket = new(fakes.FakeVersionedBucket)

		artifact = new(fakes.FakeVersionedArtifact)

		restorer = NewVersionedRestorer(map[string]VersionedBucket{
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

			dropletsBucket.CopyVersionsReturns(nil)
			buildpacksBucket.CopyVersionsReturns(nil)
			packagesBucket.CopyVersionsReturns(nil)
		})

		It("restores a backup to the corresponding buckets", func() {
			Expect(err).NotTo(HaveOccurred())

			expectedSourceRegionName, expectedSourceBucketName, expectedVersions := dropletsBucket.CopyVersionsArgsForCall(0)
			Expect(expectedSourceBucketName).To(Equal("my_droplets_bucket"))
			Expect(expectedSourceRegionName).To(Equal("my_droplets_region"))
			Expect(expectedVersions).To(Equal([]BlobVersion{
				{BlobKey: "one", Id: "13"},
				{BlobKey: "two", Id: "22"},
			}))

			expectedSourceRegionName, expectedSourceBucketName, expectedVersions = buildpacksBucket.CopyVersionsArgsForCall(0)
			Expect(expectedSourceBucketName).To(Equal("my_buildpacks_bucket"))
			Expect(expectedSourceRegionName).To(Equal("my_buildpacks_region"))
			Expect(expectedVersions).To(Equal([]BlobVersion{
				{BlobKey: "three", Id: "32"},
			}))

			expectedSourceRegionName, expectedSourceBucketName, expectedVersions = packagesBucket.CopyVersionsArgsForCall(0)
			Expect(expectedSourceBucketName).To(Equal("my_packages_bucket"))
			Expect(expectedSourceRegionName).To(Equal("my_packages_region"))
			Expect(expectedVersions).To(Equal([]BlobVersion{
				{BlobKey: "four", Id: "43"},
			}))
		})
	})

	Context("when the artifact fails to load", func() {
		BeforeEach(func() {
			artifact.LoadReturns(nil, errors.New("artifact failed to load"))
		})

		It("stops and returns an error", func() {
			Expect(err).To(MatchError("artifact failed to load"))
			Expect(dropletsBucket.CopyVersionsCallCount()).To(Equal(0))
			Expect(buildpacksBucket.CopyVersionsCallCount()).To(Equal(0))
			Expect(packagesBucket.CopyVersionsCallCount()).To(Equal(0))
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

			dropletsBucket.CopyVersionsReturns(nil)
			buildpacksBucket.CopyVersionsReturns(errors.New("failed to put versions to bucket 'buildpacks'"))
			packagesBucket.CopyVersionsReturns(nil)
		})

		It("stops and returns an error", func() {
			Expect(err).To(MatchError("failed to put versions to bucket 'buildpacks'"))

			expectedSourceRegionName, expectedSourceBucketName, expectedVersions := buildpacksBucket.CopyVersionsArgsForCall(0)
			Expect(expectedSourceBucketName).To(Equal("my_buildpacks_bucket"))
			Expect(expectedSourceRegionName).To(Equal("my_buildpacks_region"))
			Expect(expectedVersions).To(Equal([]BlobVersion{
				{BlobKey: "three", Id: "32"},
			}))
		})
	})
})
