package jobq

import (
	"context"
)

type Queue struct {
	jobs chan Job
}

func NewQueue(size int) *Queue {
	return &Queue{jobs: make(chan Job, size)}
}

func (q *Queue) Enqueue(job Job) error {
	q.jobs <- job
	return nil
}

func (q *Queue) EnqueueTask(ctx context.Context, task Task) (*Future, error) {
	job := NewJob(ctx, task)
	err := q.Enqueue(job)
	if err != nil {
		return nil, err
	}
	return job.Future, nil
}

func (q *Queue) Dequeue() (Job, error) {
	job, ok := <-q.jobs
	if !ok {
		return Job{}, ErrQueueClosed
	}
	return job, nil
}

func (q *Queue) Close() error {
	close(q.jobs)
	return nil
}

func (q *Queue) Len() int {
	return len(q.jobs)
}
