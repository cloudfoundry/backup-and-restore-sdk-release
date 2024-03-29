// Code generated by counterfeiter. DO NOT EDIT.
package fakes

import (
	gcs "gcs-blobstore-backup-restore"
	"sync"
)

type FakeBucket struct {
	CopyBlobToBucketStub        func(gcs.Bucket, string, string) error
	copyBlobToBucketMutex       sync.RWMutex
	copyBlobToBucketArgsForCall []struct {
		arg1 gcs.Bucket
		arg2 string
		arg3 string
	}
	copyBlobToBucketReturns struct {
		result1 error
	}
	copyBlobToBucketReturnsOnCall map[int]struct {
		result1 error
	}
	CopyBlobsToBucketStub        func(gcs.Bucket, string) error
	copyBlobsToBucketMutex       sync.RWMutex
	copyBlobsToBucketArgsForCall []struct {
		arg1 gcs.Bucket
		arg2 string
	}
	copyBlobsToBucketReturns struct {
		result1 error
	}
	copyBlobsToBucketReturnsOnCall map[int]struct {
		result1 error
	}
	DeleteBlobStub        func(string) error
	deleteBlobMutex       sync.RWMutex
	deleteBlobArgsForCall []struct {
		arg1 string
	}
	deleteBlobReturns struct {
		result1 error
	}
	deleteBlobReturnsOnCall map[int]struct {
		result1 error
	}
	ListBlobsStub        func(string) ([]gcs.Blob, error)
	listBlobsMutex       sync.RWMutex
	listBlobsArgsForCall []struct {
		arg1 string
	}
	listBlobsReturns struct {
		result1 []gcs.Blob
		result2 error
	}
	listBlobsReturnsOnCall map[int]struct {
		result1 []gcs.Blob
		result2 error
	}
	NameStub        func() string
	nameMutex       sync.RWMutex
	nameArgsForCall []struct {
	}
	nameReturns struct {
		result1 string
	}
	nameReturnsOnCall map[int]struct {
		result1 string
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeBucket) CopyBlobToBucket(arg1 gcs.Bucket, arg2 string, arg3 string) error {
	fake.copyBlobToBucketMutex.Lock()
	ret, specificReturn := fake.copyBlobToBucketReturnsOnCall[len(fake.copyBlobToBucketArgsForCall)]
	fake.copyBlobToBucketArgsForCall = append(fake.copyBlobToBucketArgsForCall, struct {
		arg1 gcs.Bucket
		arg2 string
		arg3 string
	}{arg1, arg2, arg3})
	stub := fake.CopyBlobToBucketStub
	fakeReturns := fake.copyBlobToBucketReturns
	fake.recordInvocation("CopyBlobToBucket", []interface{}{arg1, arg2, arg3})
	fake.copyBlobToBucketMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeBucket) CopyBlobToBucketCallCount() int {
	fake.copyBlobToBucketMutex.RLock()
	defer fake.copyBlobToBucketMutex.RUnlock()
	return len(fake.copyBlobToBucketArgsForCall)
}

func (fake *FakeBucket) CopyBlobToBucketCalls(stub func(gcs.Bucket, string, string) error) {
	fake.copyBlobToBucketMutex.Lock()
	defer fake.copyBlobToBucketMutex.Unlock()
	fake.CopyBlobToBucketStub = stub
}

func (fake *FakeBucket) CopyBlobToBucketArgsForCall(i int) (gcs.Bucket, string, string) {
	fake.copyBlobToBucketMutex.RLock()
	defer fake.copyBlobToBucketMutex.RUnlock()
	argsForCall := fake.copyBlobToBucketArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *FakeBucket) CopyBlobToBucketReturns(result1 error) {
	fake.copyBlobToBucketMutex.Lock()
	defer fake.copyBlobToBucketMutex.Unlock()
	fake.CopyBlobToBucketStub = nil
	fake.copyBlobToBucketReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeBucket) CopyBlobToBucketReturnsOnCall(i int, result1 error) {
	fake.copyBlobToBucketMutex.Lock()
	defer fake.copyBlobToBucketMutex.Unlock()
	fake.CopyBlobToBucketStub = nil
	if fake.copyBlobToBucketReturnsOnCall == nil {
		fake.copyBlobToBucketReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.copyBlobToBucketReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeBucket) CopyBlobsToBucket(arg1 gcs.Bucket, arg2 string) error {
	fake.copyBlobsToBucketMutex.Lock()
	ret, specificReturn := fake.copyBlobsToBucketReturnsOnCall[len(fake.copyBlobsToBucketArgsForCall)]
	fake.copyBlobsToBucketArgsForCall = append(fake.copyBlobsToBucketArgsForCall, struct {
		arg1 gcs.Bucket
		arg2 string
	}{arg1, arg2})
	stub := fake.CopyBlobsToBucketStub
	fakeReturns := fake.copyBlobsToBucketReturns
	fake.recordInvocation("CopyBlobsToBucket", []interface{}{arg1, arg2})
	fake.copyBlobsToBucketMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeBucket) CopyBlobsToBucketCallCount() int {
	fake.copyBlobsToBucketMutex.RLock()
	defer fake.copyBlobsToBucketMutex.RUnlock()
	return len(fake.copyBlobsToBucketArgsForCall)
}

func (fake *FakeBucket) CopyBlobsToBucketCalls(stub func(gcs.Bucket, string) error) {
	fake.copyBlobsToBucketMutex.Lock()
	defer fake.copyBlobsToBucketMutex.Unlock()
	fake.CopyBlobsToBucketStub = stub
}

func (fake *FakeBucket) CopyBlobsToBucketArgsForCall(i int) (gcs.Bucket, string) {
	fake.copyBlobsToBucketMutex.RLock()
	defer fake.copyBlobsToBucketMutex.RUnlock()
	argsForCall := fake.copyBlobsToBucketArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeBucket) CopyBlobsToBucketReturns(result1 error) {
	fake.copyBlobsToBucketMutex.Lock()
	defer fake.copyBlobsToBucketMutex.Unlock()
	fake.CopyBlobsToBucketStub = nil
	fake.copyBlobsToBucketReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeBucket) CopyBlobsToBucketReturnsOnCall(i int, result1 error) {
	fake.copyBlobsToBucketMutex.Lock()
	defer fake.copyBlobsToBucketMutex.Unlock()
	fake.CopyBlobsToBucketStub = nil
	if fake.copyBlobsToBucketReturnsOnCall == nil {
		fake.copyBlobsToBucketReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.copyBlobsToBucketReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeBucket) DeleteBlob(arg1 string) error {
	fake.deleteBlobMutex.Lock()
	ret, specificReturn := fake.deleteBlobReturnsOnCall[len(fake.deleteBlobArgsForCall)]
	fake.deleteBlobArgsForCall = append(fake.deleteBlobArgsForCall, struct {
		arg1 string
	}{arg1})
	stub := fake.DeleteBlobStub
	fakeReturns := fake.deleteBlobReturns
	fake.recordInvocation("DeleteBlob", []interface{}{arg1})
	fake.deleteBlobMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeBucket) DeleteBlobCallCount() int {
	fake.deleteBlobMutex.RLock()
	defer fake.deleteBlobMutex.RUnlock()
	return len(fake.deleteBlobArgsForCall)
}

func (fake *FakeBucket) DeleteBlobCalls(stub func(string) error) {
	fake.deleteBlobMutex.Lock()
	defer fake.deleteBlobMutex.Unlock()
	fake.DeleteBlobStub = stub
}

func (fake *FakeBucket) DeleteBlobArgsForCall(i int) string {
	fake.deleteBlobMutex.RLock()
	defer fake.deleteBlobMutex.RUnlock()
	argsForCall := fake.deleteBlobArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeBucket) DeleteBlobReturns(result1 error) {
	fake.deleteBlobMutex.Lock()
	defer fake.deleteBlobMutex.Unlock()
	fake.DeleteBlobStub = nil
	fake.deleteBlobReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeBucket) DeleteBlobReturnsOnCall(i int, result1 error) {
	fake.deleteBlobMutex.Lock()
	defer fake.deleteBlobMutex.Unlock()
	fake.DeleteBlobStub = nil
	if fake.deleteBlobReturnsOnCall == nil {
		fake.deleteBlobReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.deleteBlobReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeBucket) ListBlobs(arg1 string) ([]gcs.Blob, error) {
	fake.listBlobsMutex.Lock()
	ret, specificReturn := fake.listBlobsReturnsOnCall[len(fake.listBlobsArgsForCall)]
	fake.listBlobsArgsForCall = append(fake.listBlobsArgsForCall, struct {
		arg1 string
	}{arg1})
	stub := fake.ListBlobsStub
	fakeReturns := fake.listBlobsReturns
	fake.recordInvocation("ListBlobs", []interface{}{arg1})
	fake.listBlobsMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeBucket) ListBlobsCallCount() int {
	fake.listBlobsMutex.RLock()
	defer fake.listBlobsMutex.RUnlock()
	return len(fake.listBlobsArgsForCall)
}

func (fake *FakeBucket) ListBlobsCalls(stub func(string) ([]gcs.Blob, error)) {
	fake.listBlobsMutex.Lock()
	defer fake.listBlobsMutex.Unlock()
	fake.ListBlobsStub = stub
}

func (fake *FakeBucket) ListBlobsArgsForCall(i int) string {
	fake.listBlobsMutex.RLock()
	defer fake.listBlobsMutex.RUnlock()
	argsForCall := fake.listBlobsArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeBucket) ListBlobsReturns(result1 []gcs.Blob, result2 error) {
	fake.listBlobsMutex.Lock()
	defer fake.listBlobsMutex.Unlock()
	fake.ListBlobsStub = nil
	fake.listBlobsReturns = struct {
		result1 []gcs.Blob
		result2 error
	}{result1, result2}
}

func (fake *FakeBucket) ListBlobsReturnsOnCall(i int, result1 []gcs.Blob, result2 error) {
	fake.listBlobsMutex.Lock()
	defer fake.listBlobsMutex.Unlock()
	fake.ListBlobsStub = nil
	if fake.listBlobsReturnsOnCall == nil {
		fake.listBlobsReturnsOnCall = make(map[int]struct {
			result1 []gcs.Blob
			result2 error
		})
	}
	fake.listBlobsReturnsOnCall[i] = struct {
		result1 []gcs.Blob
		result2 error
	}{result1, result2}
}

func (fake *FakeBucket) Name() string {
	fake.nameMutex.Lock()
	ret, specificReturn := fake.nameReturnsOnCall[len(fake.nameArgsForCall)]
	fake.nameArgsForCall = append(fake.nameArgsForCall, struct {
	}{})
	stub := fake.NameStub
	fakeReturns := fake.nameReturns
	fake.recordInvocation("Name", []interface{}{})
	fake.nameMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeBucket) NameCallCount() int {
	fake.nameMutex.RLock()
	defer fake.nameMutex.RUnlock()
	return len(fake.nameArgsForCall)
}

func (fake *FakeBucket) NameCalls(stub func() string) {
	fake.nameMutex.Lock()
	defer fake.nameMutex.Unlock()
	fake.NameStub = stub
}

func (fake *FakeBucket) NameReturns(result1 string) {
	fake.nameMutex.Lock()
	defer fake.nameMutex.Unlock()
	fake.NameStub = nil
	fake.nameReturns = struct {
		result1 string
	}{result1}
}

func (fake *FakeBucket) NameReturnsOnCall(i int, result1 string) {
	fake.nameMutex.Lock()
	defer fake.nameMutex.Unlock()
	fake.NameStub = nil
	if fake.nameReturnsOnCall == nil {
		fake.nameReturnsOnCall = make(map[int]struct {
			result1 string
		})
	}
	fake.nameReturnsOnCall[i] = struct {
		result1 string
	}{result1}
}

func (fake *FakeBucket) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.copyBlobToBucketMutex.RLock()
	defer fake.copyBlobToBucketMutex.RUnlock()
	fake.copyBlobsToBucketMutex.RLock()
	defer fake.copyBlobsToBucketMutex.RUnlock()
	fake.deleteBlobMutex.RLock()
	defer fake.deleteBlobMutex.RUnlock()
	fake.listBlobsMutex.RLock()
	defer fake.listBlobsMutex.RUnlock()
	fake.nameMutex.RLock()
	defer fake.nameMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeBucket) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ gcs.Bucket = new(FakeBucket)
