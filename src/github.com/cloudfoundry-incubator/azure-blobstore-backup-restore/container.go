package azure

import (
	"fmt"
	"net/url"

	"context"

	"encoding/base64"

	"time"

	"strings"

	"github.com/Azure/azure-storage-blob-go/2017-07-29/azblob"
)

//go:generate counterfeiter -o fakes/fake_container.go . Container
type Container interface {
	Name() string
	SoftDeleteEnabled() (bool, error)
	ListBlobs() ([]BlobId, error)
	CopyBlobsFrom(containerName string, blobIds []BlobId) error
	CopyBlobsFromDifferentStorageAccount(storageAccount StorageAccount, containerName string, blobIds []BlobId) error
}

type SDKContainer struct {
	name        string
	service     azblob.ServiceURL
	environment Environment
}

func NewSDKContainer(name, storageAccount, storageKey string, environment Environment) (container SDKContainer, err error) {
	credential, err := buildCredential(storageAccount, storageKey)
	if err != nil {
		return SDKContainer{}, err
	}

	suffix, err := environment.Suffix()
	if err != nil {
		return SDKContainer{}, err
	}

	azureURL, err := url.Parse(fmt.Sprintf("https://%s.blob.%s", storageAccount, suffix))
	if err != nil {
		return SDKContainer{}, fmt.Errorf("invalid account name: '%s'", storageAccount)
	}

	service := azblob.NewServiceURL(*azureURL, azblob.NewPipeline(credential, azblob.PipelineOptions{}))
	return SDKContainer{
		name:        name,
		service:     service,
		environment: environment,
	}, nil
}

func (c SDKContainer) Name() string {
	return c.name
}

func (c SDKContainer) CopyBlobsFrom(sourceContainerName string, blobIds []BlobId) error {
	sourceContainerURL := c.service.NewContainerURL(sourceContainerName)

	return c.copyBlobs(sourceContainerName, sourceContainerURL, blobIds)
}

func (c SDKContainer) CopyBlobsFromDifferentStorageAccount(sourceStorageAccount StorageAccount, sourceContainerName string, blobIds []BlobId) error {
	suffix, err := c.environment.Suffix()
	if err != nil {
		return err
	}

	sasQueryString, err := c.buildSASQueryString(sourceContainerName, sourceStorageAccount)
	if err != nil {
		return err
	}

	sourceContainerURL, err := url.Parse(fmt.Sprintf("https://%s.blob.%s/%s?%s", sourceStorageAccount.Name, suffix, sourceContainerName, sasQueryString))
	if err != nil {
		return fmt.Errorf("invalid account name: '%s'", sourceStorageAccount.Name)
	}

	sourceContainerURLWithSAS := azblob.NewContainerURL(*sourceContainerURL, azblob.NewPipeline(azblob.NewAnonymousCredential(), azblob.PipelineOptions{}))

	return c.copyBlobs(sourceContainerName, sourceContainerURLWithSAS, blobIds)
}

func (c SDKContainer) copyBlobs(sourceContainerName string, sourceContainerURL azblob.ContainerURL, blobIds []BlobId) error {
	blobs, err := c.fetchBlobs(sourceContainerURL)
	if err != nil {
		return err
	}

	errs := make(chan error, len(blobIds))
	for _, blobId := range blobIds {
		sourceBlob, ok := blobs[blobId]
		if !ok {
			return fmt.Errorf("no \"%s\" blob with \"%s\" ETag found in container \"%s\"", blobId.Name, blobId.ETag, sourceContainerName)
		}

		go func(blob azblob.Blob) {
			errs <- c.copyBlob(sourceContainerName, sourceContainerURL, sourceBlob)
		}(sourceBlob)
	}

	var errors []error
	for range blobIds {
		err := <-errs
		if err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) != 0 {
		return formatErrors(
			fmt.Sprintf("failed to copy blob from container \"%s\" to \"%s\"", sourceContainerName, c.name),
			errors,
		)
	}

	return nil
}

func (c SDKContainer) buildSASQueryString(containerName string, storageAccount StorageAccount) (string, error) {
	sasSignatureValues := azblob.BlobSASSignatureValues{
		Protocol:   azblob.SASProtocolHTTPS,
		ExpiryTime: time.Now().Add(1 * time.Hour),
		Permissions: azblob.ContainerSASPermissions{
			List:  true,
			Read:  true,
			Write: true,
		}.String(),
		ContainerName: containerName,
	}

	sourceCredential, err := buildCredential(storageAccount.Name, storageAccount.Key)
	if err != nil {
		return "", err
	}

	sasQueryParameters := sasSignatureValues.NewSASQueryParameters(sourceCredential)

	return sasQueryParameters.Encode(), nil
}

func (c SDKContainer) fetchBlobs(sourceContainerURL azblob.ContainerURL) (map[BlobId]azblob.Blob, error) {
	var blobs = map[BlobId]azblob.Blob{}

	for marker := (azblob.Marker{}); marker.NotDone(); {
		page, err := sourceContainerURL.ListBlobsFlatSegment(
			context.Background(),
			marker,
			azblob.ListBlobsSegmentOptions{
				Details: azblob.BlobListingDetails{
					Snapshots: true,
					Deleted:   true,
				},
			},
		)
		if err != nil {
			return nil, err
		}

		marker = page.NextMarker

		for _, blob := range page.Blobs.Blob {
			blobId := BlobId{Name: blob.Name, ETag: string(blob.Properties.Etag)}
			blobs[blobId] = blob
		}
	}

	return blobs, nil
}

func (c SDKContainer) copyBlob(sourceContainerName string, sourceContainerURL azblob.ContainerURL, blob azblob.Blob) error {
	ctx := context.Background()

	sourceBlobURL := sourceContainerURL.NewBlobURL(blob.Name)
	destinationContainerURL := c.service.NewContainerURL(c.name)
	destinationBlobURL := destinationContainerURL.NewBlobURL(blob.Name)

	_, err := sourceBlobURL.Undelete(ctx)
	if err != nil {
		return err
	}

	response, err := destinationBlobURL.StartCopyFromURL(
		ctx,
		sourceBlobURL.WithSnapshot(blob.Snapshot).URL(),
		azblob.Metadata{},
		azblob.BlobAccessConditions{},
		azblob.BlobAccessConditions{},
	)
	if err != nil {
		return err
	}

	copyStatus := response.CopyStatus()

	for copyStatus == azblob.CopyStatusPending {
		time.Sleep(time.Second * 2)
		getMetadata, err := destinationBlobURL.GetProperties(ctx, azblob.BlobAccessConditions{})
		if err != nil {
			return err
		}

		copyStatus = getMetadata.CopyStatus()
	}

	if copyStatus != azblob.CopyStatusSuccess {
		return fmt.Errorf("copy of blob '%s' from container '%s' to container '%s' failed with status '%s'", blob.Name, sourceContainerName, c.Name(), copyStatus)
	}

	return nil
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

func (c SDKContainer) ListBlobs() ([]BlobId, error) {
	var blobs []BlobId
	client := c.service.NewContainerURL(c.name)

	for marker := (azblob.Marker{}); marker.NotDone(); {
		page, err := client.ListBlobsFlatSegment(context.Background(), marker, azblob.ListBlobsSegmentOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed listing blobs in container '%s': %s", c.name, err)
		}

		marker = page.NextMarker

		for _, blobInfo := range page.Blobs.Blob {
			blobs = append(blobs, BlobId{Name: blobInfo.Name, ETag: string(blobInfo.Properties.Etag)})
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

func formatErrors(contextString string, errors []error) error {
	errorStrings := make([]string, len(errors))
	for i, err := range errors {
		errorStrings[i] = err.Error()
	}
	return fmt.Errorf("%s: %s", contextString, strings.Join(errorStrings, "\n"))
}
