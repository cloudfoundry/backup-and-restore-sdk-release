package azure

import "time"

func NewTestSDKContainer(name string, storageAccount StorageAccount, environment Environment, injectedBlobLister BlobLister) (container SDKContainer, err error) {
	rtn, err := NewSDKContainer(name, storageAccount, environment)
	rtn.blobLister = injectedBlobLister
	rtn.backoffInterval = time.Duration(0)
	return rtn, err
}
