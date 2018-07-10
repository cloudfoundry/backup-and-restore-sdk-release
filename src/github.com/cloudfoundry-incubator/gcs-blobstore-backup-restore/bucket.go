package gcs

//go:generate counterfeiter -o fakes/fake_bucket.go . Bucket
type Bucket interface {
	Name() string
	VersioningEnabled() (bool, error)
	ListBlobs() ([]Blob, error)
}

func BuildBuckets(config map[string]Config) map[string]Bucket {
	buckets := map[string]Bucket{}

	for bucketId, bucketConfig := range config {
		buckets[bucketId] = NewSDKBucket(bucketConfig.Name)
	}

	return buckets
}

type Blob struct {
	Name string `json:"name"`
}

type BucketBackup struct {
	Name  string `json:"name"`
	Blobs []Blob `json:"blobs"`
}

type SDKBucket struct {
	name string
}

func NewSDKBucket(name string) SDKBucket {
	return SDKBucket{name: name}
}

func (b SDKBucket) Name() string {
	return b.name
}

func (b SDKBucket) VersioningEnabled() (bool, error) {
	return true, nil
}

func (b SDKBucket) ListBlobs() ([]Blob, error) {
	return []Blob{}, nil
}
