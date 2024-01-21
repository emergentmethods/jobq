package jobq

import (
	"sync"
)

// WorkerPool is a pool of workers that process Jobs from a JobQueue.
type WorkerPool struct {
	queue    *JobQueue
	wg       sync.WaitGroup
	shutdown chan bool
}

// NewWorkerPool creates a new WorkerPool with the given JobQueue and number of workers
// and starts the workers.
func NewWorkerPool(queue *JobQueue, numWorkers int) *WorkerPool {
	pool := &WorkerPool{
		queue:    queue,
		shutdown: make(chan bool, 1),
	}
	pool.wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go pool.worker()
	}
	return pool
}

// worker is a single worker that processes Jobs from the JobQueue.
func (p *WorkerPool) worker() {
	defer p.wg.Done()
	for {
		select {
		case <-p.shutdown:
			return
		default:
			job, err := p.queue.DequeueJob()
			if err != nil {
				// Handle errors here
				continue
			}
			job.Run()
		}
	}
}

// Close closes the WorkerPool.
func (p *WorkerPool) Close() {
	close(p.shutdown)
	p.queue.Close()
	p.wg.Wait()
}
