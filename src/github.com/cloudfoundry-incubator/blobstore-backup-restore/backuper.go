package blobstore

type Backuper struct {
	regionName                                       string
	dropletsBucket, buildpacksBucket, packagesBucket Bucket
	artifact                                         Artifact
}

func NewBackuper(regionName string, dropletsBucket, buildpacksBucket, packagesBucket Bucket, artifact Artifact) Backuper {
	return Backuper{
		regionName:       regionName,
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
		RegionName: b.regionName,
		DropletsBackup: BucketBackup{
			BucketName: b.dropletsBucket.Name(),
			Versions:   filterLatest(dropletVersions),
		},
		BuildpacksBackup: BucketBackup{
			BucketName: b.buildpacksBucket.Name(),
			Versions:   filterLatest(buildpackVersions),
		},
		PackagesBackup: BucketBackup{
			BucketName: b.packagesBucket.Name(),
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
