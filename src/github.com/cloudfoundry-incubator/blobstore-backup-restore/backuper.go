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

func (b Backuper) Backup() {
	b.artifact.Save(Backup{
		RegionName: b.regionName,
		DropletsBackup: BucketBackup{
			BucketName: b.dropletsBucket.Name(),
			Versions:   filterLatest(b.dropletsBucket.Versions()),
		},
		BuildpacksBackup: BucketBackup{
			BucketName: b.buildpacksBucket.Name(),
			Versions:   filterLatest(b.buildpacksBucket.Versions()),
		},
		PackagesBackup: BucketBackup{
			BucketName: b.packagesBucket.Name(),
			Versions:   filterLatest(b.packagesBucket.Versions()),
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
