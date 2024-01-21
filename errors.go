package jobq

import (
	"errors"
)

var (
	// ErrQueueClosed is returned when a Queue is closed.
	ErrQueueClosed = errors.New("queue closed")
	// ErrFutureClosed is returned when a Future is closed.
	ErrFutureClosed = errors.New("future closed")
	// ErrJobTimeout is returned when a Job times out.
	ErrJobTimeout = errors.New("job timeout")
)
