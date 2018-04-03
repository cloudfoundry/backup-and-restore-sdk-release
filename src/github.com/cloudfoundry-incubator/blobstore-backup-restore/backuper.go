package blobstore

import (
	"fmt"
)

//go:generate counterfeiter -o fakes/fake_clock.go . Clock
type Clock interface {
	Now() string
}

type VersionedBackuper struct {
	sourceBuckets       map[string]VersionedBucket
	destinationArtifact VersionedArtifact
}

type UnversionedBackuper struct {
	bucketPairs         map[string]UnversionedBucketPair
	clock               Clock
	destinationArtifact UnversionedArtifact
}

func (b UnversionedBackuper) Run() error {
	timestamp := b.clock.Now()
	addresses := map[string]BackupBucketAddress{}
	for id, pair := range b.bucketPairs {
		address, err := pair.Backup(timestamp + "/" + id)
		if err != nil {
			return err
		}
		addresses[id] = address
	}
	return b.destinationArtifact.Save(addresses)
}

func NewVersionedBackuper(sourceBuckets map[string]VersionedBucket, destinationArtifact VersionedArtifact) VersionedBackuper {
	return VersionedBackuper{
		sourceBuckets:       sourceBuckets,
		destinationArtifact: destinationArtifact,
	}
}

func NewUnversionedBackuper(
	bucketPairs map[string]UnversionedBucketPair,
	destinationArtifact UnversionedArtifact,
	clock Clock,
) UnversionedBackuper {
	return UnversionedBackuper{
		bucketPairs:         bucketPairs,
		destinationArtifact: destinationArtifact,
		clock:               clock,
	}
}

func (b VersionedBackuper) Run() error {
	bucketSnapshots := map[string]BucketSnapshot{}

	for identifier, bucketToBackup := range b.sourceBuckets {
		versions, err := bucketToBackup.Versions()
		if err != nil {
			return err
		}

		latestVersions := filterLatest(versions)
		if containsNullVersion(latestVersions) {
			return fmt.Errorf("failed to retrieve versions; bucket '%s' has `null` VerionIds", bucketToBackup.Name())
		}

		bucketSnapshots[identifier] = BucketSnapshot{
			BucketName: bucketToBackup.Name(),
			RegionName: bucketToBackup.RegionName(),
			Versions:   latestVersions,
		}
	}

	return b.destinationArtifact.Save(bucketSnapshots)
}

func containsNullVersion(latestVersions []BlobVersion) bool {
	for _, version := range latestVersions {
		if version.Id == "null" {
			return true
		}
	}
	return false
}

func filterLatest(versions []Version) []BlobVersion {
	var filteredVersions []BlobVersion
	for _, version := range versions {
		if version.IsLatest {
			filteredVersions = append(filteredVersions, BlobVersion{Id: version.Id, BlobKey: version.Key})
		}
	}
	return filteredVersions
}
