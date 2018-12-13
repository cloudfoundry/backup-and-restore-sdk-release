package gcs

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

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
	ListBlobs(prefix string) ([]Blob, error)
	ListBackups() ([]string, error)
	CopyBlobWithinBucket(src string, dst string) error
	CopyBlobToBucket(bucket Bucket, src string, dst string) error
	CopyBlobsToBucket(bucket Bucket, src string) error
	DeleteBlob(name string) error
	MarkBackupComplete(prefix string) error
	IsBackupComplete(prefix string) (bool, error)
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

func (b SDKBucket) ListBlobs(prefix string) ([]Blob, error) {
	var blobs []Blob

	query := &storage.Query{
		Prefix: prefix,
	}
	objectsIterator := b.handle.Objects(b.ctx, query)

	for {
		objectAttributes, err := objectsIterator.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}

			return nil, err
		}

		blobs = append(blobs, NewBlob(objectAttributes.Name))
	}

	return blobs, nil
}

func (b SDKBucket) ListBackups() ([]string, error) {
	var dirs []string

	storageQuery := &storage.Query{
		Delimiter: "/",
	}
	objectsIterator := b.handle.Objects(b.ctx, storageQuery)
	for {
		objectAttributes, err := objectsIterator.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}

			return nil, err
		}

		dir := strings.TrimSuffix(objectAttributes.Prefix, "/")
		regex := regexp.MustCompile(`^\d{4}(_\d{2}){5}$`)

		if regex.MatchString(dir) {
			dirs = append(dirs, dir)
		}
	}

	return dirs, nil
}

func (b SDKBucket) CopyBlobWithinBucket(srcBlob, dstBlob string) error {
	return b.CopyBlobToBucket(b, srcBlob, dstBlob)
}

func (b SDKBucket) CopyBlobToBucket(dstBucket Bucket, srcBlob, dstBlob string) error {
	if dstBucket == nil {
		return errors.New("destination bucket does not exist")
	}

	src := b.client.Bucket(b.Name()).Object(srcBlob)
	dst := b.client.Bucket(dstBucket.Name()).Object(dstBlob)
	_, err := dst.CopierFrom(src).Run(b.ctx)
	if err != nil {
		return fmt.Errorf("failed to copy object: %v", err)
	}

	return nil
}

func (b SDKBucket) CopyBlobsToBucket(dstBucket Bucket, srcPrefix string) error {
	if dstBucket == nil {
		return errors.New("destination bucket does not exist")
	}

	blobs, err := b.ListBlobs(srcPrefix)
	if err != nil {
		return err
	}

	for _, blob := range blobs {
		if blob.IsBackupComplete() {
			continue
		}

		destinationName := strings.TrimPrefix(blob.Name(), srcPrefix+blobNameDelimiter)

		err = b.CopyBlobToBucket(dstBucket, blob.Name(), destinationName)
		if err != nil {
			return err
		}
	}

	return nil
}

func (b SDKBucket) DeleteBlob(blob string) error {
	return b.client.Bucket(b.Name()).Object(blob).Delete(b.ctx)
}

func (b SDKBucket) MarkBackupComplete(prefix string) error {
	blob := NewBackupCompleteBlob(prefix)
	writer := b.client.Bucket(b.Name()).Object(blob.Name()).NewWriter(b.ctx)

	_, err := writer.Write([]byte{})
	if err != nil {
		return fmt.Errorf("failed creating backup complete blob: %s", err)
	}

	err = writer.Close()
	if err != nil {
		return fmt.Errorf("failed creating backup complete blob: %s", err)
	}

	return nil
}

func (b SDKBucket) IsBackupComplete(prefix string) (bool, error) {
	blob := NewBackupCompleteBlob(prefix)
	_, err := b.handle.Object(blob.Name()).Attrs(b.ctx)
	if err != nil {
		if err == storage.ErrObjectNotExist {
			return false, nil
		}

		return false, fmt.Errorf("failed checking backup complete blob: %s", err)
	}

	return true, nil
}
