package azure

import (
	"fmt"
	"net/url"

	"context"

	"github.com/Azure/azure-storage-blob-go/2017-07-29/azblob"
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

func NewContainer(containerConfig ContainerConfig) (SDKContainer, error) {
	credential := azblob.NewSharedKeyCredential(containerConfig.AzureAccountName, containerConfig.AzureAccountKey)
	pipeline := azblob.NewPipeline(credential, azblob.PipelineOptions{})
	azureURL, err := url.Parse(fmt.Sprintf("https://%s.blob.core.windows.net", containerConfig.AzureAccountName))
	if err != nil {
		return SDKContainer{}, err
	}

	serviceURL := azblob.NewServiceURL(*azureURL, pipeline)
	return SDKContainer{
		name:   containerConfig.Name,
		client: serviceURL.NewContainerURL(containerConfig.Name),
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
			return nil, err
		}

		marker = page.NextMarker

		for _, blobInfo := range page.Blobs.Blob {
			blobs = append(blobs, Blob{Name: blobInfo.Name, Hash: *blobInfo.Properties.ContentMD5})
		}
	}

	return blobs, nil
}
