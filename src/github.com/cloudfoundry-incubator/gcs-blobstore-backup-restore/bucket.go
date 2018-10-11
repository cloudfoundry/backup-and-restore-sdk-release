package gcs

import (
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
	ListBlobs() ([]Blob, error)
	CopyBlob(string, string) (int64, error)
}

type BucketPair struct {
	Bucket       Bucket
	BackupBucket Bucket
}

func BuildBuckets(config map[string]Config) (map[string]BucketPair, error) {
	buckets := map[string]BucketPair{}

	for bucketId, bucketConfig := range config {
		bucket, err := NewSDKBucket(bucketConfig.ServiceAccountKey, bucketConfig.BucketName)
		if err != nil {
			return nil, err
		}

		backupBucket, err := NewSDKBucket(bucketConfig.ServiceAccountKey, bucketConfig.BackupBucketName)
		if err != nil {
			return nil, err
		}

		buckets[bucketId] = BucketPair{
			Bucket:       bucket,
			BackupBucket: backupBucket,
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

func (b SDKBucket) CopyBlob(sourceBlob string, newBlob string) (int64, error) {

	src := b.client.Bucket(b.name).Object(sourceBlob)
	dst := b.client.Bucket(b.name).Object(newBlob)

	attr, err := dst.CopierFrom(src).Run(b.ctx)

	if err != nil {
		return 0, err
	}

	return attr.Generation, nil
}
