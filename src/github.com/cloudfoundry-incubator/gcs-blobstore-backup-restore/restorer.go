package gcs

type Restorer struct {
	buckets           map[string]BucketPair
	executionStrategy Strategy
}

func NewRestorer(buckets map[string]BucketPair, executionStrategy Strategy) Restorer {
	return Restorer{
		buckets:           buckets,
		executionStrategy: executionStrategy,
	}
}

func (r Restorer) Restore(backups map[string]BucketBackup) error {
	//for bucketIdentifier := range backups {
	//	_, exists := r.buckets[bucketIdentifier]
	//	if !exists {
	//		return fmt.Errorf("bucket identifier '%s' not found in buckets configuration", bucketIdentifier)
	//	}
	//}
	//
	//for _, bucket := range r.buckets {
	//	enabled, err := bucket.VersioningEnabled()
	//	if err != nil {
	//		return fmt.Errorf("failed to check if versioning is enabled on bucket '%s': %s", bucket.Name(), err)
	//	}
	//
	//	if !enabled {
	//		return fmt.Errorf("versioning is not enabled on bucket '%s'", bucket.Name())
	//	}
	//}
	//
	//for bucketIdentifier, backup := range backups {
	//	bucket := r.buckets[bucketIdentifier]
	//
	//	errs := r.executionStrategy.Run(backup.Blobs, func(blob Blob) error {
	//		return bucket.CopyVersion(blob, backup.Name)
	//	})
	//
	//	if len(errs) != 0 {
	//		return formatErrors(fmt.Sprintf("failed to restore bucket '%s'", bucket.Name()), errs)
	//	}
	//}

	return nil
}

//func formatErrors(contextString string, errors []error) error {
//	errorStrings := make([]string, len(errors))
//	for i, err := range errors {
//		errorStrings[i] = err.Error()
//	}
//	return fmt.Errorf("%s: %s", contextString, strings.Join(errorStrings, "\n"))
//}
