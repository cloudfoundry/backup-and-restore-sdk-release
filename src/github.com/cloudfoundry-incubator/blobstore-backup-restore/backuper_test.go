package blobstore_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry-incubator/blobstore-backup-restore"
	"github.com/cloudfoundry-incubator/blobstore-backup-restore/fakes"
)

var _ = Describe("Backuper", func() {
	var dropletsBucket *fakes.FakeBucket
	var buildpacksBucket *fakes.FakeBucket
	var packagesBucket *fakes.FakeBucket

	var artifact *fakes.FakeArtifact

	var backuper Backuper

	BeforeEach(func() {
		dropletsBucket = new(fakes.FakeBucket)
		buildpacksBucket = new(fakes.FakeBucket)
		packagesBucket = new(fakes.FakeBucket)

		artifact = new(fakes.FakeArtifact)

		backuper = NewBackuper("eu-west-1", dropletsBucket, buildpacksBucket, packagesBucket, artifact)
	})

	Context("The buckets have data", func() {
		BeforeEach(func() {
			dropletsBucket.VersionsReturns([]Version{
				{Key: "one", Id: "11", IsLatest: false},
				{Key: "one", Id: "12", IsLatest: false},
				{Key: "one", Id: "13", IsLatest: true},
				{Key: "two", Id: "21", IsLatest: false},
				{Key: "two", Id: "22", IsLatest: true},
			})
			dropletsBucket.NameReturns("my_droplets_bucket")

			buildpacksBucket.VersionsReturns([]Version{
				{Key: "three", Id: "31", IsLatest: false},
				{Key: "three", Id: "32", IsLatest: true},
			})
			buildpacksBucket.NameReturns("my_buildpacks_bucket")

			packagesBucket.VersionsReturns([]Version{
				{Key: "four", Id: "41", IsLatest: false},
				{Key: "four", Id: "43", IsLatest: true},
				{Key: "four", Id: "42", IsLatest: false},
			})
			packagesBucket.NameReturns("my_packages_bucket")
		})

		JustBeforeEach(func() {
			backuper.Backup()
		})

		It("stores the latest versions in the artifact", func() {
			Expect(artifact.SaveArgsForCall(0)).To(Equal(Backup{
				RegionName: "eu-west-1",
				DropletsBackup: BucketBackup{
					BucketName: "my_droplets_bucket",
					Versions: []LatestVersion{
						{BlobKey: "one", Id: "13"},
						{BlobKey: "two", Id: "22"},
					},
				},
				BuildpacksBackup: BucketBackup{
					BucketName: "my_buildpacks_bucket",
					Versions: []LatestVersion{
						{BlobKey: "three", Id: "32"},
					},
				},
				PackagesBackup: BucketBackup{
					BucketName: "my_packages_bucket",
					Versions: []LatestVersion{
						{BlobKey: "four", Id: "43"},
					},
				},
			}))
		})
	})
})
