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

		credential := azblob.NewSharedKeyCredential(containerConfig.AzureAccountName, containerConfig.AzureAccountKey)
		pipeline := azblob.NewPipeline(credential, azblob.PipelineOptions{})
		azureURL, err := url.Parse(fmt.Sprintf("https://%s.blob.core.windows.net", containerConfig.AzureAccountName))
		if err != nil {
			return nil, err
		}

		serviceURL := azblob.NewServiceURL(*azureURL, pipeline)
		ctx := context.Background()
		containerURL := serviceURL.NewContainerURL(containerConfig.Name)

		for marker := (azblob.Marker{}); marker.NotDone(); {
			listBlob, err := containerURL.ListBlobs(ctx, marker, azblob.ListBlobsOptions{})
			if err != nil {
				return nil, err
			}

			marker = listBlob.NextMarker

			for _, blobInfo := range listBlob.Blobs.Blob {
				blobs = append(blobs, Blob{Name: blobInfo.Name, Hash: *blobInfo.Properties.ContentMD5})
			}
		}

		backups[containerId] = ContainerBackup{Name: containerConfig.Name, Blobs: blobs}
	}

	return backups, nil
}
