package gcs

type Strategy interface {
	Run([]Blob, func(blob Blob) error) []error
}

type SerialStrategy struct{}

func NewSerialStrategy() SerialStrategy {
	return SerialStrategy{}
}

func (s SerialStrategy) Run(blobs []Blob, action func(blob Blob) error) []error {
	var errs []error
	for _, blob := range blobs {
		err := action(blob)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errs
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
