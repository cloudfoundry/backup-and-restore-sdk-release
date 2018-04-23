package azure

import (
	"context"
	"fmt"
	"net/url"

	"github.com/Azure/azure-storage-blob-go/2017-07-29/azblob"
)

type Blob struct {
	Name string `json:"name"`
	Hash string `json:"hash"`
}

type ContainerBackup struct {
	Name  string `json:"name"`
	Blobs []Blob `json:"blobs"`
}

type Backuper struct {
	config map[string]ContainerConfig
}

func NewBackuper(config map[string]ContainerConfig) Backuper {
	return Backuper{config: config}
}

func (b Backuper) Backup() (map[string]ContainerBackup, error) {
	var backups = make(map[string]ContainerBackup)

	for containerId, containerConfig := range b.config {
		var blobs []Blob

		container, err := NewContainer(containerConfig)
		if err != nil {
			return nil, err
		}

		err = container.forEachBlob(func(blobInfo azblob.Blob) {
			blobs = append(blobs, Blob{Name: blobInfo.Name, Hash: *blobInfo.Properties.ContentMD5})
		})
		if err != nil {
			return nil, err
		}

		backups[containerId] = ContainerBackup{Name: containerConfig.Name, Blobs: blobs}
	}

	return backups, nil
}

type Container struct {
	client azblob.ContainerURL
}

func NewContainer(containerConfig ContainerConfig) (Container, error) {
	credential := azblob.NewSharedKeyCredential(containerConfig.AzureAccountName, containerConfig.AzureAccountKey)
	pipeline := azblob.NewPipeline(credential, azblob.PipelineOptions{})
	azureURL, err := url.Parse(fmt.Sprintf("https://%s.blob.core.windows.net", containerConfig.AzureAccountName))
	if err != nil {
		return Container{}, err
	}

	serviceURL := azblob.NewServiceURL(*azureURL, pipeline)
	return Container{client: serviceURL.NewContainerURL(containerConfig.Name)}, nil
}

func (c Container) forEachBlob(action func(blob azblob.Blob)) error {
	for marker := (azblob.Marker{}); marker.NotDone(); {
		page, err := c.client.ListBlobs(context.Background(), marker, azblob.ListBlobsOptions{})
		if err != nil {
			return err
		}

		marker = page.NextMarker

		for _, blobInfo := range page.Blobs.Blob {
			action(blobInfo)
		}
	}

	return nil
}
