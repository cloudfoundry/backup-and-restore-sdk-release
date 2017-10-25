package blobstore

import "fmt"

type Backuper struct {
	buckets  map[string]Bucket
	artifact Artifact
}

func NewBackuper(buckets map[string]Bucket, artifact Artifact) Backuper {
	return Backuper{
		buckets:  buckets,
		artifact: artifact,
	}
}

func (b Backuper) Backup() error {
	backup := map[string]BucketBackup{}

	for identifier, bucket := range b.buckets {
		versions, err := bucket.Versions()
		if err != nil {
			return err
		}

		latestVersions := filterLatest(versions)
		if containsNullVersion(latestVersions) {
			return fmt.Errorf("failed to retrieve versions; bucket '%s' has `null` VerionIds", bucket.Name())
		}

		backup[identifier] = BucketBackup{
			BucketName: bucket.Name(),
			RegionName: bucket.RegionName(),
			Versions:   latestVersions,
		}
	}

	return b.artifact.Save(backup)
}

func containsNullVersion(latestVersions []LatestVersion) bool {
	for _, version := range latestVersions {
		if version.Id == "null" {
			return true
		}
	}
	return false
}

func filterLatest(versions []Version) []LatestVersion {
	filteredVersions := []LatestVersion{}
	for _, version := range versions {
		if version.IsLatest {
			filteredVersions = append(filteredVersions, LatestVersion{Id: version.Id, BlobKey: version.Key})
		}
	}
	return filteredVersions
}
