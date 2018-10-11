package gcs

import "fmt"

type Backuper struct {
	buckets map[string]BucketPair
}

func NewBackuper(buckets map[string]BucketPair) Backuper {
	return Backuper{
		buckets: buckets,
	}
}

func (b *Backuper) CreateLiveBucketSnapshot() error {
	for _, bucketPair := range b.buckets {
		bucket := bucketPair.Bucket
		blobs, err := bucket.ListBlobs()

		if err != nil {
			return err
		}

		for _, blob := range blobs {
			_, err := bucket.CopyBlob(blob.Name, fmt.Sprintf("temporary-backup-artifact/%s", blob.Name))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (b *Backuper) Backup() (map[string]BucketBackup, error) {
	bucketBackups := map[string]BucketBackup{}

	//for _, bucket := range b.buckets {
	//	enabled, err := bucket.VersioningEnabled()
	//	if err != nil {
	//		return nil, err
	//	}
	//
	//	if !enabled {
	//		return nil, fmt.Errorf("versioning is not enabled on bucket: %s", bucket.Name())
	//	}
	//}
	//
	//for bucketIdentifier, bucket := range b.buckets {
	//	blobs, err := bucket.ListBlobs()
	//	if err != nil {
	//		return nil, err
	//	}
	//
	//	bucketBackups[bucketIdentifier] = BucketBackup{
	//		Name:  bucket.Name(),
	//		Blobs: blobs,
	//	}
	//}

	return bucketBackups, nil
}
