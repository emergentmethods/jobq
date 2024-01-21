package jobq

import (
	"errors"
)

var (
	// ErrQueueClosed is returned when a Queue is closed.
	ErrQueueClosed = errors.New("queue closed")
)
