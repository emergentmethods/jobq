package jobq

import (
	"sync"
)

type WorkerPool struct {
	queue    *Queue
	wg       sync.WaitGroup
	shutdown chan struct{}
}

func NewWorkerPool(queue *Queue, numWorkers int) *WorkerPool {
	pool := &WorkerPool{
		queue:    queue,
		shutdown: make(chan struct{}),
	}
	pool.wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go pool.worker()
	}
	return pool
}

func (p *WorkerPool) worker() {
	defer p.wg.Done()
	for {
		select {
		case <-p.shutdown:
			return
		default:
			job, err := p.queue.Dequeue()
			if err != nil {
				// Handle errors here
				continue
			}
			job.Run()
		}
	}
}

func (p *WorkerPool) Close() {
	close(p.shutdown)
	p.queue.Close()
	p.wg.Wait()
}
