package blobstore

type Backuper struct {
	dropletsBucket, buildpacksBucket, packagesBucket Bucket
	artifact                                         Artifact
}

func NewBackuper(dropletsBucket, buildpacksBucket, packagesBucket Bucket, artifact Artifact) Backuper {
	return Backuper{
		dropletsBucket:   dropletsBucket,
		buildpacksBucket: buildpacksBucket,
		packagesBucket:   packagesBucket,
		artifact:         artifact,
	}
}

func (b Backuper) Backup() error {
	dropletVersions, err := b.dropletsBucket.Versions()
	if err != nil {
		return err
	}

	buildpackVersions, err := b.buildpacksBucket.Versions()
	if err != nil {
		return err
	}

	packageVersions, err := b.packagesBucket.Versions()
	if err != nil {
		return err
	}

	return b.artifact.Save(Backup{
		DropletsBackup: BucketBackup{
			BucketName: b.dropletsBucket.Name(),
			RegionName: b.dropletsBucket.RegionName(),
			Versions:   filterLatest(dropletVersions),
		},
		BuildpacksBackup: BucketBackup{
			BucketName: b.buildpacksBucket.Name(),
			RegionName: b.buildpacksBucket.RegionName(),
			Versions:   filterLatest(buildpackVersions),
		},
		PackagesBackup: BucketBackup{
			BucketName: b.packagesBucket.Name(),
			RegionName: b.packagesBucket.RegionName(),
			Versions:   filterLatest(packageVersions),
		},
	})
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
