package blobstore

type Backuper struct {
	buckets  []Bucket
	artifact Artifact
}

func NewBackuper(buckets []Bucket, artifact Artifact) Backuper {
	return Backuper{
		buckets:  buckets,
		artifact: artifact,
	}
}

func (b Backuper) Backup() error {
	backup := map[string]BucketBackup{}

	for _, bucket := range b.buckets {
		versions, err := bucket.Versions()
		if err != nil {
			return err
		}

		backup[bucket.Identifier()] = BucketBackup{
			BucketName: bucket.Name(),
			RegionName: bucket.RegionName(),
			Versions:   filterLatest(versions),
		}
	}

	return b.artifact.Save(backup)
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
