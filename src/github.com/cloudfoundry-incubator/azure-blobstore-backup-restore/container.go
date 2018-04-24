package azure

import (
	"fmt"
	"net/url"

	"context"

	"github.com/Azure/azure-storage-blob-go/2017-07-29/azblob"
	"encoding/base64"
)

//go:generate counterfeiter -o fakes/fake_container.go . Container
type Container interface {
	Name() string
	ListBlobs() ([]Blob, error)
}

type SDKContainer struct {
	name   string
	client azblob.ContainerURL
}

func NewContainer(name, storageAccount, storageKey string) (container SDKContainer, err error) {
	credential, err := buildCredential(storageAccount, storageKey)
	if err != nil {
		return SDKContainer{}, err
	}

	pipeline := azblob.NewPipeline(credential, azblob.PipelineOptions{})
	azureURL, err := url.Parse(fmt.Sprintf("https://%s.blob.core.windows.net", storageAccount))
	if err != nil {
		return SDKContainer{}, fmt.Errorf("invalid account name: '%s'", storageAccount)
	}

	serviceURL := azblob.NewServiceURL(*azureURL, pipeline)
	return SDKContainer{
		name:   name,
		client: serviceURL.NewContainerURL(name),
	}, nil
}

func (c SDKContainer) Name() string {
	return c.name
}

func (c SDKContainer) ListBlobs() ([]Blob, error) {
	var blobs []Blob

	for marker := (azblob.Marker{}); marker.NotDone(); {
		page, err := c.client.ListBlobs(context.Background(), marker, azblob.ListBlobsOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed listing blobs in container '%s': %s", c.name, err)
		}

		marker = page.NextMarker

		for _, blobInfo := range page.Blobs.Blob {
			blobs = append(blobs, Blob{Name: blobInfo.Name, Hash: *blobInfo.Properties.ContentMD5})
		}
	}

	return blobs, nil
}

func buildCredential(storageAccount, storageKey string) (*azblob.SharedKeyCredential, error) {
	_, err := base64.StdEncoding.DecodeString(storageKey)
	if err != nil {
		return nil, fmt.Errorf("invalid storage key: '%s'", err)
	}

	return azblob.NewSharedKeyCredential(storageAccount, storageKey), nil
}
