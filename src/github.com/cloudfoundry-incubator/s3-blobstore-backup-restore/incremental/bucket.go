package incremental

//go:generate counterfeiter -o fakes/fake_bucket.go . Bucket
type Bucket interface {
	Name() string
	ListBlobs(path string) ([]Blob, error)
	ListDirectories() ([]string, error)
	CopyBlobWithinBucket(src, dst string) error
	CopyBlobFromBucket(bucket Bucket, src, dst string) error
	UploadBlob(path, contents string) error
	HasBlob(path string) (bool, error)
	IsBackupComplete(prefix string) (bool, error)
}

//go:generate counterfeiter -o fakes/fake_blob.go . Blob
type Blob interface {
	Path() string
}
