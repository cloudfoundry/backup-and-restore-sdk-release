package blobstore_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"errors"

	. "github.com/cloudfoundry-incubator/blobstore-backup-restore"
	"github.com/cloudfoundry-incubator/blobstore-backup-restore/fakes"
)

var _ = Describe("VersionedBackuper", func() {
	var dropletsBucket *fakes.FakeVersionedBucket
	var buildpacksBucket *fakes.FakeVersionedBucket
	var packagesBucket *fakes.FakeVersionedBucket

	var artifact *fakes.FakeVersionedArtifact

	var err error

	var backuper VersionedBackuper

	BeforeEach(func() {
		dropletsBucket = new(fakes.FakeVersionedBucket)
		buildpacksBucket = new(fakes.FakeVersionedBucket)
		packagesBucket = new(fakes.FakeVersionedBucket)

		artifact = new(fakes.FakeVersionedArtifact)

		backuper = NewVersionedBackuper(map[string]VersionedBucket{
			"droplets":   dropletsBucket,
			"buildpacks": buildpacksBucket,
			"packages":   packagesBucket,
		}, artifact)
	})

	JustBeforeEach(func() {
		err = backuper.Run()
	})

	Context("when the buckets have data", func() {
		BeforeEach(func() {
			dropletsBucket.NameReturns("my_droplets_bucket")
			dropletsBucket.RegionNameReturns("my_droplets_region")
			dropletsBucket.VersionsReturns([]Version{
				{Key: "one", Id: "11", IsLatest: false},
				{Key: "one", Id: "12", IsLatest: false},
				{Key: "one", Id: "13", IsLatest: true},
				{Key: "two", Id: "21", IsLatest: false},
				{Key: "two", Id: "22", IsLatest: true},
			}, nil)

			buildpacksBucket.NameReturns("my_buildpacks_bucket")
			buildpacksBucket.RegionNameReturns("my_buildpacks_region")
			buildpacksBucket.VersionsReturns([]Version{
				{Key: "three", Id: "31", IsLatest: false},
				{Key: "three", Id: "32", IsLatest: true},
			}, nil)

			packagesBucket.NameReturns("my_packages_bucket")
			packagesBucket.RegionNameReturns("my_packages_region")
			packagesBucket.VersionsReturns([]Version{
				{Key: "four", Id: "41", IsLatest: false},
				{Key: "four", Id: "43", IsLatest: true},
				{Key: "four", Id: "42", IsLatest: false},
			}, nil)
		})

		It("stores the latest versions in the artifact", func() {
			Expect(artifact.SaveArgsForCall(0)).To(Equal(map[string]BucketSnapshot{
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
			}))
		})
	})

	Context("when retrieving the versions from the buckets fails", func() {
		BeforeEach(func() {
			dropletsBucket.VersionsReturns([]Version{}, nil)
			dropletsBucket.NameReturns("my_droplets_bucket")

			buildpacksBucket.VersionsReturns([]Version{}, errors.New("failed to retrieve versions"))
			buildpacksBucket.NameReturns("my_buildpacks_bucket")

			packagesBucket.VersionsReturns([]Version{}, nil)
			packagesBucket.NameReturns("my_packages_bucket")
		})

		It("returns the error from the bucket", func() {
			Expect(err).To(MatchError("failed to retrieve versions"))
		})
	})

	Context("when there is a `null` VersionId", func() {
		Context("when it's a latest version", func() {
			BeforeEach(func() {
				packagesBucket.VersionsReturns([]Version{
					{Key: "one", Id: "11", IsLatest: true},
					{Key: "one", Id: "12", IsLatest: false},
					{Key: "two", Id: "null", IsLatest: true},
					{Key: "two", Id: "21", IsLatest: false},
				}, nil)
				packagesBucket.NameReturns("my_packages_bucket")
			})

			It("returns the error", func() {
				Expect(err).To(MatchError("failed to retrieve versions; bucket 'my_packages_bucket' has `null` VerionIds"))
			})
		})

		Context("when it's not a latest version", func() {
			BeforeEach(func() {
				packagesBucket.VersionsReturns([]Version{
					{Key: "one", Id: "11", IsLatest: true},
					{Key: "one", Id: "12", IsLatest: false},
					{Key: "two", Id: "21", IsLatest: true},
					{Key: "two", Id: "null", IsLatest: false},
				}, nil)
				packagesBucket.NameReturns("my_packages_bucket")
			})

			It("does not return an error", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Context("when storing the versions in the artifact fails", func() {
		BeforeEach(func() {
			dropletsBucket.VersionsReturns([]Version{}, nil)
			dropletsBucket.NameReturns("my_droplets_bucket")

			buildpacksBucket.VersionsReturns([]Version{}, nil)
			buildpacksBucket.NameReturns("my_buildpacks_bucket")

			packagesBucket.VersionsReturns([]Version{}, nil)
			packagesBucket.NameReturns("my_packages_bucket")

			artifact.SaveReturns(errors.New("failed to save the versions artifact"))
		})

		It("returns the error from the artifact", func() {
			Expect(err).To(MatchError("failed to save the versions artifact"))
		})
	})
})
