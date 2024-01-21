package jobq

type Future struct {
	result chan result
}

type result struct {
	value interface{}
	err   error
}

func NewFuture() *Future {
	return &Future{result: make(chan result)}
}

func (f *Future) Result() (interface{}, error) {
	r := <-f.result
	return r.value, r.err
}

func (f *Future) SetResult(value interface{}, err error) {
	f.result <- result{value, err}
}
