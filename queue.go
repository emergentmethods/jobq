package jobq

import (
	"container/heap"
	"sync"
)

// Queue is an interface for a queue that stores jobs.
type Queue interface {
	Enqueue(*Job, interface{})
	Dequeue() (*Job, error)
	Len() int
	Close()
}

// FIFOQueue is a queue that stores jobs in a FIFO order.
type FIFOQueue struct {
	jobs chan *Job
	mu   sync.Mutex
}

// NewFIFOQueue creates a new FIFOQueue with the given size.
func NewFIFOQueue(size int) *FIFOQueue {
	return &FIFOQueue{jobs: make(chan *Job, size)}
}

// Enqueue adds a job to the queue.
func (q *FIFOQueue) Enqueue(job *Job, opts interface{}) {
	// Ignore opts for FIFOQueue
	q.jobs <- job
}

// Dequeue removes a job from the queue.
func (q *FIFOQueue) Dequeue() (*Job, error) {
	job, ok := <-q.jobs
	if !ok {
		return nil, ErrQueueClosed
	}
	return job, nil
}

// Len returns the number of jobs in the queue.
func (q *FIFOQueue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.jobs)
}

// Close closes the queue.
func (q *FIFOQueue) Close() {
	q.mu.Lock()
	defer q.mu.Unlock()
	close(q.jobs)
}

type LIFOQueue struct {
	jobs   []*Job
	mu     sync.Mutex
	cond   *sync.Cond
	max    int
	closed bool
}

// NewLIFOQueue creates a new bounded LIFOQueue with the given size.
func NewLIFOQueue(size int) *LIFOQueue {
	q := &LIFOQueue{
		jobs: make([]*Job, 0, size),
		max:  size,
	}
	q.cond = sync.NewCond(&q.mu)
	return q
}

// Enqueue adds a job to the queue.
func (q *LIFOQueue) Enqueue(job *Job, opts interface{}) {
	q.mu.Lock()
	defer q.mu.Unlock()

	for len(q.jobs) == q.max && !q.closed {
		q.cond.Wait()
	}

	if q.closed {
		panic("sending on closed queue")
	}

	q.jobs = append(q.jobs, job)
	q.cond.Signal()
}

// Dequeue removes and returns the most recently added job from the queue.
func (q *LIFOQueue) Dequeue() (*Job, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	for len(q.jobs) == 0 {
		if q.closed {
			return nil, ErrQueueClosed
		}
		q.cond.Wait()
	}

	job := q.jobs[len(q.jobs)-1]
	q.jobs = q.jobs[:len(q.jobs)-1]
	q.cond.Signal()
	return job, nil
}

// Len returns the number of jobs in the queue.
func (q *LIFOQueue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.jobs)
}

// Close closes the queue.
func (q *LIFOQueue) Close() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.closed = true
	q.cond.Broadcast()
}

// PriorityQueueItem is an item in a priority queue.
type PriorityQueueItem struct {
	Job      *Job
	Priority int
	Index    int
}

// PriorityQueueHeap is a heap implementation of a priority queue.
type PriorityQueueHeap []*PriorityQueueItem

// Len returns the length of the heap.
func (h PriorityQueueHeap) Len() int {
	return len(h)
}

// Less returns true if the item at index i has a lower priority than the item at index j.
func (h PriorityQueueHeap) Less(i, j int) bool {
	return h[i].Priority < h[j].Priority
}

// Swap swaps the items at index i and j.
func (h PriorityQueueHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].Index = i
	h[j].Index = j
}

// Push pushes an item onto the heap.
func (h *PriorityQueueHeap) Push(x interface{}) {
	item := x.(*PriorityQueueItem)
	item.Index = len(*h)
	*h = append(*h, item)
}

// Pop pops an item from the heap.
func (h *PriorityQueueHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	item.Index = -1
	*h = old[0 : n-1]
	return item
}

type PriorityQueueOptions struct {
	Priority int
}

// PriorityQueue is a queue that stores jobs in a priority order.
type PriorityQueue struct {
	jobs   PriorityQueueHeap
	mu     sync.Mutex
	cond   *sync.Cond
	closed bool
	max    int
}

// NewPriorityQueue creates a new PriorityQueue with the given size.
func NewPriorityQueue(size int) *PriorityQueue {
	q := &PriorityQueue{
		jobs: make(PriorityQueueHeap, 0, size),
		max:  size,
	}
	q.cond = sync.NewCond(&q.mu)
	heap.Init(&q.jobs)
	return q
}

// Enqueue adds a job to the queue.
func (q *PriorityQueue) Enqueue(job *Job, opts interface{}) {
	q.mu.Lock()
	defer q.mu.Unlock()

	for q.jobs.Len() == q.max && !q.closed {
		q.cond.Wait()
	}

	if q.closed {
		panic("sending on closed queue")
	}

	// Set default priority
	priority := 1

	// Handle specific options for PriorityQueue
	if po, ok := opts.(*PriorityQueueOptions); ok {
		priority = po.Priority
	}

	item := &PriorityQueueItem{Job: job, Priority: priority}
	heap.Push(&q.jobs, item)
}

// Dequeue removes a job from the queue.
func (q *PriorityQueue) Dequeue() (*Job, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	for q.jobs.Len() == 0 {
		if q.closed {
			return nil, ErrQueueClosed
		}
		q.cond.Wait()
	}

	item := heap.Pop(&q.jobs).(*PriorityQueueItem)
	return item.Job, nil
}

// Len returns the number of jobs in the queue.
func (q *PriorityQueue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.jobs.Len()
}

// Close closes the queue.
func (q *PriorityQueue) Close() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.closed = true
	q.cond.Broadcast()
}
