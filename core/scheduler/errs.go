package scheduler

import "errors"

var (
	ErrJobNotFound      = errors.New("job not found")
	ErrJobAlreadyExists = errors.New("job already exists")
	ErrJobNameEmpty     = errors.New("job name is empty")
)
