package azure

import (
	"fmt"
	"net/url"

	"context"

	"encoding/base64"

	"github.com/Azure/azure-storage-blob-go/2017-07-29/azblob"
)

//go:generate counterfeiter -o fakes/fake_container.go . Container
type Container interface {
	Name() string
	SoftDeleteEnabled() (bool, error)
	ListBlobs() ([]Blob, error)
}

type SDKContainer struct {
	name    string
	service azblob.ServiceURL
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

	service := azblob.NewServiceURL(*azureURL, pipeline)
	return SDKContainer{
		name:    name,
		service: service,
	}, nil
}

func (c SDKContainer) Name() string {
	return c.name
}

func (c SDKContainer) SoftDeleteEnabled() (bool, error) {
	properties, err := c.service.GetProperties(context.Background())
	if err != nil {
		return false, fmt.Errorf("failed fetching properties for storage account: '%s'", err)
	}

	if properties.DeleteRetentionPolicy == nil {
		return true, nil
	}

	return properties.DeleteRetentionPolicy.Enabled, nil
}

func (c SDKContainer) ListBlobs() ([]Blob, error) {
	var blobs []Blob
	client := c.service.NewContainerURL(c.name)

	for marker := (azblob.Marker{}); marker.NotDone(); {
		page, err := client.ListBlobs(context.Background(), marker, azblob.ListBlobsOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed listing blobs in container '%s': %s", c.name, err)
		}

		marker = page.NextMarker

		for _, blobInfo := range page.Blobs.Blob {
			blobs = append(blobs, Blob{Name: blobInfo.Name, Etag: string(blobInfo.Properties.Etag)})
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
