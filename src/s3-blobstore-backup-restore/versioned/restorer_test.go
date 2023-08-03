package versioned_test

import (
	"errors"
	"fmt"

	"s3-blobstore-backup-restore/versioned/fakes"

	"s3-blobstore-backup-restore/versioned"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Restorer", func() {
	var dropletsBucket *fakes.FakeBucket
	var buildpacksBucket *fakes.FakeBucket
	var packagesBucket *fakes.FakeBucket

	var artifact *fakes.FakeArtifact

	var err error

	var restorer versioned.Restorer

	BeforeEach(func() {
		dropletsBucket = new(fakes.FakeBucket)
		buildpacksBucket = new(fakes.FakeBucket)
		packagesBucket = new(fakes.FakeBucket)

		dropletsBucket.IsVersionedReturns(true, nil)
		buildpacksBucket.IsVersionedReturns(true, nil)
		packagesBucket.IsVersionedReturns(true, nil)

		artifact = new(fakes.FakeArtifact)

		restorer = versioned.NewRestorer(map[string]versioned.Bucket{
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
			artifact.LoadReturns(map[string]versioned.BucketSnapshot{
				"droplets": {
					BucketName: "my_droplets_bucket",
					RegionName: "my_droplets_source_region",
					Versions: []versioned.BlobVersion{
						{BlobKey: "one", Id: "13"},
						{BlobKey: "two", Id: "22"},
					},
				},
				"buildpacks": {
					BucketName: "my_buildpacks_bucket",
					RegionName: "my_buildpacks_source_region",
					Versions: []versioned.BlobVersion{
						{BlobKey: "three", Id: "32"},
					},
				},
				"packages": {
					BucketName: "my_packages_bucket",
					RegionName: "my_packages_source_region",
					Versions: []versioned.BlobVersion{
						{BlobKey: "four", Id: "43"},
					},
				},
			}, nil)

			dropletsBucket.CopyVersionReturns(nil)
			buildpacksBucket.CopyVersionReturns(nil)
			packagesBucket.CopyVersionReturns(nil)
			dropletsBucket.RegionReturns("destination_droplets_region")
			buildpacksBucket.RegionReturns("destination_buildpacks_region")
			packagesBucket.RegionReturns("destination_packages_region")
		})

		It("restores a backup from one region to a new foundation in a different region", func() {
			By("successfully running", func() {
				Expect(err).NotTo(HaveOccurred())
			})

			By("Checking the buckets are versioned", func() {
				Expect(dropletsBucket.IsVersionedCallCount()).To(Equal(1))
				Expect(buildpacksBucket.IsVersionedCallCount()).To(Equal(1))
				Expect(packagesBucket.IsVersionedCallCount()).To(Equal(1))
			})

			By("Calling CopyVersion for each object in the droplets bucket with the old region", func() {
				Expect(dropletsBucket.CopyVersionCallCount()).To(Equal(2))

				expectedBlobKey, expectedVersionId, expectedSourceBucketName, expectedSourceRegionName := dropletsBucket.CopyVersionArgsForCall(0)
				Expect(expectedBlobKey).To(Equal("one"))
				Expect(expectedVersionId).To(Equal("13"))
				Expect(expectedSourceBucketName).To(Equal("my_droplets_bucket"))
				Expect(expectedSourceRegionName).To(Equal("my_droplets_source_region"))

				expectedBlobKey, expectedVersionId, expectedSourceBucketName, expectedSourceRegionName = dropletsBucket.CopyVersionArgsForCall(1)
				Expect(expectedBlobKey).To(Equal("two"))
				Expect(expectedVersionId).To(Equal("22"))
				Expect(expectedSourceBucketName).To(Equal("my_droplets_bucket"))
				Expect(expectedSourceRegionName).To(Equal("my_droplets_source_region"))
			})

			By("Calling CopyVersions for each object in the buildpacks bucket with the old region", func() {
				Expect(buildpacksBucket.CopyVersionCallCount()).To(Equal(1))

				expectedBlobKey, expectedVersionId, expectedSourceBucketName, expectedSourceRegionName := buildpacksBucket.CopyVersionArgsForCall(0)
				Expect(expectedBlobKey).To(Equal("three"))
				Expect(expectedVersionId).To(Equal("32"))
				Expect(expectedSourceBucketName).To(Equal("my_buildpacks_bucket"))
				Expect(expectedSourceRegionName).To(Equal("my_buildpacks_source_region"))
			})

			By("Calling CopyVersions for each object in the packages bucket with the old region", func() {
				Expect(packagesBucket.CopyVersionCallCount()).To(Equal(1))

				expectedBlobKey, expectedVersionId, expectedSourceBucketName, expectedSourceRegionName := packagesBucket.CopyVersionArgsForCall(0)
				Expect(expectedBlobKey).To(Equal("four"))
				Expect(expectedVersionId).To(Equal("43"))
				Expect(expectedSourceBucketName).To(Equal("my_packages_bucket"))
				Expect(expectedSourceRegionName).To(Equal("my_packages_source_region"))
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
			artifact.LoadReturns(map[string]versioned.BucketSnapshot{
				"droplets": {
					BucketName: "my_droplets_bucket",
					RegionName: "my_droplets_region",
					Versions: []versioned.BlobVersion{
						{BlobKey: "one", Id: "13"},
						{BlobKey: "two", Id: "22"},
					},
				},
				"buildpacks": {
					BucketName: "my_buildpacks_bucket",
					RegionName: "my_buildpacks_region",
					Versions: []versioned.BlobVersion{
						{BlobKey: "three", Id: "32"},
					},
				},
				"packages": {
					BucketName: "my_packages_bucket",
					RegionName: "my_packages_region",
					Versions: []versioned.BlobVersion{
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

	Context("when the bucket is not versioned", func() {
		BeforeEach(func() {
			artifact.LoadReturns(map[string]versioned.BucketSnapshot{
				"droplets": {
					BucketName: "my_droplets_bucket",
					RegionName: "my_droplets_region",
					Versions: []versioned.BlobVersion{
						{BlobKey: "one", Id: "13"},
						{BlobKey: "two", Id: "22"},
					},
				},
				"buildpacks": {
					BucketName: "my_buildpacks_bucket",
					RegionName: "my_buildpacks_region",
					Versions: []versioned.BlobVersion{
						{BlobKey: "three", Id: "32"},
					},
				},
				"packages": {
					BucketName: "my_packages_bucket",
					RegionName: "my_packages_region",
					Versions: []versioned.BlobVersion{
						{BlobKey: "four", Id: "43"},
					},
				},
			}, nil)

			dropletsBucket.CopyVersionReturns(nil)
			buildpacksBucket.CopyVersionReturns(nil)
			packagesBucket.CopyVersionReturns(nil)

			dropletsBucket.IsVersionedReturns(false, nil)
			dropletsBucket.NameReturns("my_droplets_bucket")
		})

		It("returns an error", func() {
			Expect(err).To(MatchError(fmt.Errorf("bucket my_droplets_bucket is not versioned")))
		})
	})

	Context("when it fails to check if the bucket is versioned", func() {
		BeforeEach(func() {
			artifact.LoadReturns(map[string]versioned.BucketSnapshot{
				"droplets": {
					BucketName: "my_droplets_bucket",
					RegionName: "my_droplets_region",
					Versions: []versioned.BlobVersion{
						{BlobKey: "one", Id: "13"},
						{BlobKey: "two", Id: "22"},
					},
				},
				"buildpacks": {
					BucketName: "my_buildpacks_bucket",
					RegionName: "my_buildpacks_region",
					Versions: []versioned.BlobVersion{
						{BlobKey: "three", Id: "32"},
					},
				},
				"packages": {
					BucketName: "my_packages_bucket",
					RegionName: "my_packages_region",
					Versions: []versioned.BlobVersion{
						{BlobKey: "four", Id: "43"},
					},
				},
			}, nil)

			dropletsBucket.CopyVersionReturns(nil)
			buildpacksBucket.CopyVersionReturns(nil)
			packagesBucket.CopyVersionReturns(nil)

			dropletsBucket.IsVersionedReturns(false, fmt.Errorf("ooops"))
			dropletsBucket.NameReturns("my_droplets_bucket")
		})

		It("returns an error", func() {
			Expect(err).To(MatchError(fmt.Errorf("failed to check if my_droplets_bucket is versioned: ooops")))
		})
	})

	Context("when there isn't a corresponding bucket recorded in backup artifact", func() {
		BeforeEach(func() {
			artifact.LoadReturns(map[string]versioned.BucketSnapshot{
				"droplets": {
					BucketName: "my_droplets_bucket",
					RegionName: "my_droplets_region",
					Versions: []versioned.BlobVersion{
						{BlobKey: "one", Id: "13"},
						{BlobKey: "two", Id: "22"},
					},
				},
				"buildpacks": {
					BucketName: "my_buildpacks_bucket",
					RegionName: "my_buildpacks_region",
					Versions: []versioned.BlobVersion{
						{BlobKey: "three", Id: "32"},
					},
				},
			}, nil)

			restorer = versioned.NewRestorer(map[string]versioned.Bucket{
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

	Context("when there is a bucket recorded in backup artifact but not in the restore config", func() {
		BeforeEach(func() {
			artifact.LoadReturns(map[string]versioned.BucketSnapshot{
				"droplets": {
					BucketName: "my_droplets_bucket",
					RegionName: "my_droplets_region",
					Versions: []versioned.BlobVersion{
						{BlobKey: "one", Id: "13"},
						{BlobKey: "two", Id: "22"},
					},
				},
				"buildpacks": {
					BucketName: "my_buildpacks_bucket",
					RegionName: "my_buildpacks_region",
					Versions: []versioned.BlobVersion{
						{BlobKey: "three", Id: "32"},
					},
				},
			}, nil)

			restorer = versioned.NewRestorer(map[string]versioned.Bucket{
				"droplets": dropletsBucket,
			}, artifact)
		})

		It("fails and returns a useful error", func() {
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("no entry found in restore config for bucket: buildpacks"))
		})
	})
})
