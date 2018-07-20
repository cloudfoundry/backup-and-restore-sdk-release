package gcs

import (
	"fmt"

	"cloud.google.com/go/storage"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

const readWriteScope = "https://www.googleapis.com/auth/devstorage.read_write"

//go:generate counterfeiter -o fakes/fake_bucket.go . Bucket
type Bucket interface {
	Name() string
	VersioningEnabled() (bool, error)
	ListBlobs() ([]Blob, error)
	CopyVersion(blob Blob, sourceBucketName string) error
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
	client *storage.Client
}

func NewSDKBucket(serviceAccountKeyJson string, name string) (SDKBucket, error) {
	ctx := context.Background()

	creds, err := google.CredentialsFromJSON(ctx, []byte(serviceAccountKeyJson), readWriteScope)
	if err != nil {
		return SDKBucket{}, err
	}

	client, err := storage.NewClient(ctx, option.WithCredentials(creds))
	if err != nil {
		return SDKBucket{}, err
	}

	handle := client.Bucket(name)

	return SDKBucket{name: name, handle: handle, ctx: ctx, client: client}, nil
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

func (b SDKBucket) CopyVersion(blob Blob, sourceBucketName string) error {
	ctx := context.Background()

	sourceObjectHandle := b.client.Bucket(sourceBucketName).Object(blob.Name)
	_, err := sourceObjectHandle.Generation(blob.GenerationID).Attrs(ctx)
	if err != nil {
		return fmt.Errorf("error getting blob version attributes 'gs://%s/%s#%d': %s", sourceBucketName, blob.Name, blob.GenerationID, err)
	}

	if b.name == sourceBucketName {
		attrs, err := b.handle.Object(blob.Name).Attrs(ctx)

		if err == nil && attrs.Generation == blob.GenerationID {
			return nil
		}
	}

	source := sourceObjectHandle.Generation(blob.GenerationID)
	copier := b.handle.Object(blob.Name).CopierFrom(source)
	_, err = copier.Run(ctx)
	if err != nil {
		return fmt.Errorf("error copying blob 'gs://%s/%s#%d': %s", sourceBucketName, blob.Name, blob.GenerationID, err)
	}

	return nil
}
