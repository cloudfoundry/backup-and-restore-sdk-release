package gcs

type Strategy interface {
	Run([]Blob, func(blob Blob) error) []error
}

type ParallelStrategy struct{}

func NewParallelStrategy() ParallelStrategy {
	return ParallelStrategy{}
}

func (s ParallelStrategy) Run(blobs []Blob, action func(blob Blob) error) []error {
	var errors []error
	errs := make(chan error, len(blobs))

	for _, blob := range blobs {
		go func(blob Blob) {
			errs <- action(blob)
		}(blob)
	}

	for range blobs {
		err := <-errs
		if err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}
