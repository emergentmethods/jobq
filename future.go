package jobq

import "time"

var futureTimeout = 10 * time.Second

// Future is a future result of a Job.
type Future struct {
	result chan result
	closed bool
}

// result is the result of a Job.
type result struct {
	value interface{}
	err   error
}

// NewFuture creates a new Future.
func NewFuture() *Future {
	// Note: The buffer size of 1 is important here. If the buffer size is 0, the result
	// will be sent to the result channel and block until the result is read. If the
	// buffer size is 1, the result will be sent to the result channel and the Job will
	// continue executing.
	return &Future{result: make(chan result, 1), closed: false}
}

// Result returns the result of the Job. This method blocks until the result is available.
func (f *Future) Result() (interface{}, error) {
	// r := <-f.result
	// return r.value, r.err
	if f.closed {
		return nil, ErrFutureClosed
	}

	r := <-f.result
	f.Release()
	return r.value, r.err
}

// SetResult sets the result of the Job.
func (f *Future) SetResult(value interface{}, err error) {
	select {
	case f.result <- result{value, err}:
		return
	case <-time.After(futureTimeout):
		f.Release()
	}
}

// Release releases the memory used by the Future.
func (f *Future) Release() {
	if !f.closed {
		f.closed = true
		close(f.result)
	}
}
