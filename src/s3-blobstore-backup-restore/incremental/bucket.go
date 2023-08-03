package incremental

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate -o fakes/fake_bucket.go . Bucket
type Bucket interface {
	Name() string
	Region() string
	ListBlobs(path string) ([]Blob, error)
	ListDirectories() ([]string, error)
	CopyBlobWithinBucket(src, dst string) error
	CopyBlobFromBucket(bucket Bucket, src, dst string) error
	UploadBlob(path, contents string) error
	HasBlob(path string) (bool, error)
}

//counterfeiter:generate -o fakes/fake_blob.go . Blob
type Blob interface {
	Path() string
}
