package jobq

import (
	"errors"
)

var (
	ErrQueueClosed = errors.New("queue closed")
)
