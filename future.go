package jobq

// Future is a future result of a Job.
type Future struct {
	result chan result
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
	return &Future{result: make(chan result, 1)}
}

// Result returns the result of the Job. This method blocks until the result is available.
func (f *Future) Result() (interface{}, error) {
	r := <-f.result
	return r.value, r.err
}

// SetResult sets the result of the Job.
func (f *Future) SetResult(value interface{}, err error) {
	f.result <- result{value, err}
}
