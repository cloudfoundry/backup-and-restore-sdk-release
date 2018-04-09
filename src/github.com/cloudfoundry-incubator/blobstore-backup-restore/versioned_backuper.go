package blobstore

import (
	"fmt"

	"github.com/cloudfoundry-incubator/blobstore-backup-restore/s3"
)

type VersionedBackuper struct {
	sourceBuckets       map[string]s3.VersionedBucket
	destinationArtifact VersionedArtifact
}

func NewVersionedBackuper(sourceBuckets map[string]s3.VersionedBucket, destinationArtifact VersionedArtifact) VersionedBackuper {
	return VersionedBackuper{
		sourceBuckets:       sourceBuckets,
		destinationArtifact: destinationArtifact,
	}
}

func (b VersionedBackuper) Run() error {
	bucketSnapshots := map[string]BucketSnapshot{}

	for identifier, bucketToBackup := range b.sourceBuckets {
		versions, err := bucketToBackup.ListVersions()
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

func filterLatest(versions []s3.Version) []BlobVersion {
	var filteredVersions []BlobVersion
	for _, version := range versions {
		if version.IsLatest {
			filteredVersions = append(filteredVersions, BlobVersion{Id: version.Id, BlobKey: version.Key})
		}
	}
	return filteredVersions
}
