package azure

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"

	"time"

	oldblob "github.com/Azure/azure-storage-blob-go/azblob"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/sas"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/service"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate -o fakes/fake_container.go . Container
type Container interface {
	Name() string
	URL() string
	SoftDeleteEnabled() (bool, error)
	ListBlobs() ([]BlobId, error)
	CopyBlobsFromSameStorageAccount(containerName string, blobIds []BlobId) error
	CopyBlobsFromDifferentStorageAccount(storageAccount StorageAccount, containerName string, blobIds []BlobId) error
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate -o fakes/fake_blob_lister.go . BlobLister
type BlobLister interface {
	ListBlobsFlatSegment(ctx context.Context, marker oldblob.Marker, o oldblob.ListBlobsSegmentOptions) (*oldblob.ListBlobsFlatSegmentResponse, error)
}

type SDKContainer struct {
	name            string
	containerClient *container.Client
	serviceClient   *service.Client
	environment     Environment
}

func buildServiceClient(storageAccount StorageAccount, environment Environment) (*service.Client, error) {
	creds, err := buildCredential(storageAccount)
	if err != nil {
		return nil, err
	}

	suffix, err := environment.Suffix()
	if err != nil {
		return nil, err
	}

	serviceURL, err := url.Parse(fmt.Sprintf("https://%s.blob.%s", storageAccount.Name, suffix))
	if err != nil {
		return nil, fmt.Errorf("invalid account name: '%s'", storageAccount.Name)
	}

	serviceClient, err := service.NewClientWithSharedKeyCredential(serviceURL.String(), creds, nil)
	if err != nil {
		return nil, err
	}

	return serviceClient, nil
}

func NewSDKContainer(name string, storageAccount StorageAccount, environment Environment) (SDKContainer, error) {

	serviceClient, err := buildServiceClient(storageAccount, environment)
	if err != nil {
		return SDKContainer{}, err
	}

	containerClient := serviceClient.NewContainerClient(name)

	return SDKContainer{
		name:            name,
		containerClient: containerClient,
		serviceClient:   serviceClient,
		environment:     environment,
	}, nil
}

func (c SDKContainer) Name() string {
	return c.name
}

func (c SDKContainer) URL() string {
	return c.containerClient.URL()
}

func (c SDKContainer) CopyBlobsFromSameStorageAccount(sourceContainerName string, blobIds []BlobId) error {
	sourceContainerClient := c.serviceClient.NewContainerClient(sourceContainerName)

	return c.copyBlobs(sourceContainerClient, blobIds)
}

func (c SDKContainer) CopyBlobsFromDifferentStorageAccount(sourceStorageAccount StorageAccount, sourceContainerName string, blobIds []BlobId) error {
	sourceServiceClient, err := buildServiceClient(sourceStorageAccount, c.environment)
	if err != nil {
		return err
	}

	sourceContainerClient := sourceServiceClient.NewContainerClient(sourceContainerName)

	containerPermissions := sas.ContainerPermissions{
		List:  true,
		Read:  true,
		Write: true,
	}

	urlExpiryTime := time.Now().Add(1 * time.Hour)
	// On 2024-02-26 we used sourceContainerClient.GetSASURL to create a
	// "Shared Access Signature" URL. This URL gives the holder permission
	// to access the container in question for a given length of time.
	sourceContainerSASURL, err := sourceContainerClient.GetSASURL(containerPermissions, urlExpiryTime, nil)
	if err != nil {
		return err
	}

	sourceContainerClientWithSAS, err := container.NewClientWithNoCredential(sourceContainerSASURL, nil)
	if err != nil {
		return err
	}

	return c.copyBlobs(sourceContainerClientWithSAS, blobIds)
}

func (c SDKContainer) copyBlobs(sourceContainerClient *container.Client, blobIds []BlobId) error {
	blobs, err := c.fetchBlobs(sourceContainerClient)
	if err != nil {
		return err
	}

	errs := make(chan error, len(blobIds))
	for _, blobId := range blobIds {
		sourceBlob, ok := blobs[blobId]
		if !ok {
			return fmt.Errorf("no \"%s\" blob with \"%s\" ETag found in container \"%s\"", blobId.Name, blobId.ETag, sourceContainerClient.URL())
		}

		if sourceBlob == nil {
			return fmt.Errorf("nil sourceBlob")
		}

		go func(blob *container.BlobItem) {
			errs <- c.copyBlob(sourceContainerClient, blob)
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
			fmt.Sprintf("failed to copy blob from container \"%s\" to \"%s\"", sourceContainerClient.URL(), c.name),
			errors,
		)
	}

	return nil
}

func (c SDKContainer) copyBlob(sourceContainerClient *container.Client, blobItem *container.BlobItem) error {

	ctx := context.Background()

	destinationBlobClient := c.containerClient.NewBlobClient(*blobItem.Name)
	sourceBlobClient := sourceContainerClient.NewBlobClient(*blobItem.Name)

	_, err := sourceBlobClient.Undelete(ctx, nil) // UndeleteOptions
	if err != nil {
		return err
	}

	sourceURL := sourceBlobClient.URL()

	if blobItem.Snapshot != nil {
		sourceBlobClientWithSnapshot, err := sourceBlobClient.WithSnapshot(*blobItem.Snapshot)
		if err != nil {
			return err
		}

		sourceURL = sourceBlobClientWithSnapshot.URL()
	}

	resp, err := destinationBlobClient.StartCopyFromURL(ctx, sourceURL, nil)
	if err != nil {
		return err
	}

	copyStatus := *resp.CopyStatus

	for copyStatus == blob.CopyStatusTypePending {
		time.Sleep(time.Second * 2)
		getMetadata, err := destinationBlobClient.GetProperties(ctx, nil)
		if err != nil {
			return err
		}

		copyStatus = *getMetadata.CopyStatus
	}

	if copyStatus != blob.CopyStatusTypeSuccess {
		return fmt.Errorf("copy of blob '%s' from container '%s' to container '%s' failed with status '%s'", blobItem.Name, sourceContainerClient.URL(), c.Name(), copyStatus)
	}

	return nil
}

func (c SDKContainer) fetchBlobs(sourceContainerClient *container.Client) (map[BlobId]*container.BlobItem, error) {

	blobs := make(map[BlobId]*container.BlobItem)
	pager := sourceContainerClient.NewListBlobsFlatPager(&container.ListBlobsFlatOptions{
		Include: container.ListBlobsInclude{
			Deleted:   true,
			Snapshots: true,
		},
	})

	for pager.More() {
		page, err := pager.NextPage(context.Background())
		if err != nil {
			return nil, fmt.Errorf("failed listing blobs in container '%s': %s", c.name, err)
		}

		for _, blobItem := range page.ListBlobsFlatSegmentResponse.Segment.BlobItems {
			blobs[BlobId{Name: *blobItem.Name, ETag: string(*blobItem.Properties.ETag)}] = blobItem
		}
	}

	return blobs, nil
}

func (c SDKContainer) SoftDeleteEnabled() (bool, error) {
	properties, err := c.serviceClient.GetProperties(context.Background(), nil)
	if err != nil {
		return false, fmt.Errorf("failed fetching properties for storage account: '%s'", err)
	}

	if properties.DeleteRetentionPolicy == nil {
		return true, nil
	}

	return *properties.DeleteRetentionPolicy.Enabled, nil
}

func (c SDKContainer) ListBlobs() ([]BlobId, error) {
	var blobs []BlobId

	pager := c.containerClient.NewListBlobsFlatPager(nil)

	for pager.More() {
		page, err := pager.NextPage(context.Background())
		if err != nil {
			return nil, fmt.Errorf("failed listing blobs in container '%s': %s", c.name, err)
		}

		for _, blobInfo := range page.ListBlobsFlatSegmentResponse.Segment.BlobItems {
			blobs = append(blobs, BlobId{Name: *blobInfo.Name, ETag: string(*blobInfo.Properties.ETag)})
		}
	}

	return blobs, nil
}

func buildCredential(storageAccount StorageAccount) (*azblob.SharedKeyCredential, error) {
	_, err := base64.StdEncoding.DecodeString(storageAccount.Key)
	if err != nil {
		return nil, fmt.Errorf("invalid storage key: '%s'", err)
	}

	return azblob.NewSharedKeyCredential(storageAccount.Name, storageAccount.Key)
}

func formatErrors(contextString string, errors []error) error {
	errorStrings := make([]string, len(errors))
	for i, err := range errors {
		errorStrings[i] = err.Error()
	}
	return fmt.Errorf("%s: %s", contextString, strings.Join(errorStrings, "\n"))
}
