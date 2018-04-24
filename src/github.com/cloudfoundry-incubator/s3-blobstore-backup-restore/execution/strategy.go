package execution

type Strategy interface {
	Run([]string, func(file string) error) []error
}

type SerialStrategy struct{}

func NewSerialStrategy() SerialStrategy {
	return SerialStrategy{}
}

func (s SerialStrategy) Run(files []string, action func(string) error) []error {
	var errs []error
	for _, file := range files {
		err := action(file)
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

func (s ParallelStrategy) Run(files []string, action func(string) error) []error {
	var errors []error
	errs := make(chan error, len(files))

	for _, file := range files {
		go func(file string) {
			errs <- action(file)
		}(file)
	}

	for range files {
		err := <-errs
		if err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}
