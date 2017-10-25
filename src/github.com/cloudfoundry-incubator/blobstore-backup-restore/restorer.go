package blobstore

type Restorer struct {
	buckets  map[string]Bucket
	artifact Artifact
}

func NewRestorer(buckets map[string]Bucket, artifact Artifact) Restorer {
	return Restorer{buckets: buckets, artifact: artifact}
}

func (r Restorer) Restore() error {
	backup, err := r.artifact.Load()
	if err != nil {
		return err
	}

	for identifier, bucket := range r.buckets {
		bucketBackup := backup[identifier]
		err = bucket.PutVersions(bucketBackup.RegionName, bucketBackup.BucketName, bucketBackup.Versions)
		if err != nil {
			return err
		}
	}

	return nil
}
