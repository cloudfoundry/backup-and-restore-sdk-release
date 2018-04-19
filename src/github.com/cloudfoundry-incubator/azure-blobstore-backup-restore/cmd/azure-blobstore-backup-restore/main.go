package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"

	"context"

	"github.com/Azure/azure-storage-blob-go/2017-07-29/azblob"
)

type ContainerConfig struct {
	Name             string `json:"name"`
	AzureAccountName string `json:"azure_account_name"`
	AzureAccountKey  string `json:"azure_account_key"`
}

type Blob struct {
	Name string `json:"name"`
	Hash string `json:"hash"`
}

type Artifact struct {
	Name  string `json:"name"`
	Blobs []Blob `json:"blobs"`
}

func main() {
	var artifactFilePath = flag.String("artifact-file", "", "Path to the artifact file")
	var configFilePath = flag.String("config", "", "Path to JSON config file")
	var _ = flag.Bool("backup", false, "Run blobstore backup")

	flag.Parse()

	config, err := parseConfig(configFilePath)
	if err != nil {
		log.Fatalf(err.Error())
	}

	artifact, err := takeBackup(config)
	if err != nil {
		log.Fatalf(err.Error())
	}

	writeArtifact(artifact, artifactFilePath)
	if err != nil {
		log.Fatalf(err.Error())
	}
}

func parseConfig(configFilePath *string) (map[string]ContainerConfig, error) {
	configContents, err := ioutil.ReadFile(*configFilePath)
	if err != nil {
		return nil, err
	}

	var config map[string]ContainerConfig
	err = json.Unmarshal(configContents, &config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func takeBackup(config map[string]ContainerConfig) (map[string]Artifact, error) {
	var artifact = make(map[string]Artifact)

	for containerId, containerConfig := range config {
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

		artifact[containerId] = Artifact{Name: containerConfig.Name, Blobs: blobs}
	}

	return artifact, nil
}

func writeArtifact(artifact map[string]Artifact, artifactFilePath *string) error {
	filesContents, err := json.Marshal(artifact)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(*artifactFilePath, filesContents, 0644)
}
