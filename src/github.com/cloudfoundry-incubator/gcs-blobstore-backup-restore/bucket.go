package gcs

import (
	"cloud.google.com/go/storage"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

const readOnlyScope = "https://www.googleapis.com/auth/devstorage.read_only"

//go:generate counterfeiter -o fakes/fake_bucket.go . Bucket
type Bucket interface {
	Name() string
	VersioningEnabled() (bool, error)
	ListBlobs() ([]Blob, error)
}

func BuildBuckets(config map[string]Config) (map[string]Bucket, error) {
	buckets := map[string]Bucket{}

	var err error
	for bucketId, bucketConfig := range config {
		buckets[bucketId], err = NewSDKBucket(bucketConfig.ServiceAccountKey, bucketConfig.Name)
		if err != nil {
			return nil, err
		}
	}

	return buckets, nil
}

type Blob struct {
	Name         string `json:"name"`
	GenerationID int64  `json:"generation_id"`
}

type BucketBackup struct {
	Name  string `json:"name"`
	Blobs []Blob `json:"blobs"`
}

type SDKBucket struct {
	name   string
	handle *storage.BucketHandle
	ctx    context.Context
}

func NewSDKBucket(serviceAccountKeyJson string, name string) (SDKBucket, error) {
	ctx := context.Background()

	creds, err := google.CredentialsFromJSON(ctx, []byte(serviceAccountKeyJson), readOnlyScope)
	if err != nil {
		return SDKBucket{}, err
	}

	client, err := storage.NewClient(ctx, option.WithCredentials(creds))
	if err != nil {
		return SDKBucket{}, err
	}

	handle := client.Bucket(name)

	return SDKBucket{name: name, handle: handle, ctx: ctx}, nil
}

func (b SDKBucket) Name() string {
	return b.name
}

func (b SDKBucket) VersioningEnabled() (bool, error) {
	attrs, err := b.handle.Attrs(b.ctx)
	if err != nil {
		return false, err
	}

	return attrs.VersioningEnabled, nil
}

func (b SDKBucket) ListBlobs() ([]Blob, error) {
	var blobs []Blob

	objectsIterator := b.handle.Objects(b.ctx, nil)
	for {
		objectAttributes, err := objectsIterator.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			return nil, err
		}

		blobs = append(blobs, Blob{Name: objectAttributes.Name, GenerationID: objectAttributes.Generation})
	}

	return blobs, nil
}
