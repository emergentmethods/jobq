package jobq

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Task is a common interface for tasks that can be executed by a Job.
type Task interface {
	Execute(context.Context) (interface{}, error)
}

// JobStatus is the status of a Job.
type JobStatus string

const (
	// StatusPending is the status of a Job that has not yet been enqueued.
	StatusPending JobStatus = "pending"
	// StatusRunning is the status of a Job that is currently running.
	StatusRunning JobStatus = "running"
	// StatusCompleted is the status of a Job that has completed successfully.
	StatusCompleted JobStatus = "completed"
	// StatusFailed is the status of a Job that has failed.
	StatusFailed JobStatus = "failed"
)

var (
	// jobMaxTimeout is the maximum amount of time a Job can run before it is cancelled.
	jobMaxTimeout = 10 * time.Minute
	// jobStatusMap is a map of Job IDs to JobStatuses.
	jobStatusMap = make(map[uuid.UUID]JobStatus)
	mapMutex     sync.Mutex
)

// Job is a unit of work that can be executed by a Worker.
type Job struct {
	ID           uuid.UUID
	Task         Task
	Ctx          context.Context
	Future       *Future
	QueueOptions interface{}
	maxRetries   int
	retries      int
}

// JobOptions are options for creating a new Job.
type JobOptions struct {
	ID           uuid.UUID
	Task         Task
	Ctx          context.Context
	MaxRetries   int
	QueueOptions interface{}
}

// NewJob creates a new Job with the given Task.
func NewJob(opts *JobOptions) (*Job, error) {
	// We set maxRetries to 1 if it is less than or equal to 0
	// to ensure the task runs at least once.
	if opts.MaxRetries <= 0 {
		opts.MaxRetries = 1
	}
	if opts.ID == uuid.Nil {
		opts.ID = uuid.New()
	}
	if opts.Ctx == nil {
		opts.Ctx = context.Background()
	}
	if opts.Task == nil {
		return nil, errors.New("Task is required")
	}

	job := &Job{
		Ctx:          opts.Ctx,
		Task:         opts.Task,
		Future:       NewFuture(),
		ID:           opts.ID,
		QueueOptions: opts.QueueOptions,
		maxRetries:   opts.MaxRetries,
		retries:      0,
	}
	job.SetStatus(StatusPending)
	return job, nil
}

// executeJob executes the Job's Task, returning the result and error.
// If the Job's context is cancelled, the context error is returned.
func (j *Job) executeJob() (interface{}, error) {
	done := make(chan bool, 1)

	var result interface{}
	var err error

	go func() {
		defer close(done)
		result, err = j.Task.Execute(j.Ctx)
	}()

	select {
	case <-j.Ctx.Done():
		return nil, j.Ctx.Err()
	case <-done:
		return result, err
	case <-time.After(jobMaxTimeout):
		return nil, ErrJobTimeout
	}
}

// executeWithRetries executes the Job's Task with retries.
func (j *Job) executeWithRetries() bool {
	var err error

	for j.retries < j.maxRetries {
		result, execErr := j.executeJob()

		if execErr != nil {
			err = execErr
			j.retries++
			// TODO: exponential backoff?
		} else {
			j.Future.SetResult(result, nil)
			return true
		}
	}

	j.Future.SetResult(nil, err)
	return false
}

// Run executes the Job's Task.
func (j *Job) Run() {
	j.SetStatus(StatusRunning)

	if j.executeWithRetries() {
		j.SetStatus(StatusCompleted)
	} else {
		j.SetStatus(StatusFailed)
	}
}

// Result returns the result of the Job's Task.
func (j *Job) Result() (interface{}, error) {
	return j.Future.Result()
}

// Empty returns true if the Job is empty.
func (j *Job) Empty() bool {
	return j.Task == nil
}

// SetStatus sets the status of the Job.
func (j *Job) SetStatus(status JobStatus) {
	mapMutex.Lock()
	jobStatusMap[j.ID] = status
	if status == StatusCompleted || status == StatusFailed {
		delete(jobStatusMap, j.ID)
	}
	mapMutex.Unlock()
}

// GetStatus returns the status of the Job.
func (j *Job) GetStatus() JobStatus {
	mapMutex.Lock()
	status := jobStatusMap[j.ID]
	mapMutex.Unlock()
	return status
}

// JobQueue is a FIFO queue for Jobs.
type JobQueue struct {
	queue Queue
}

// JobQueueOptions are options for creating a new JobQueue.
type JobQueueOptions struct {
	Queue Queue
}

// NewJobQueue creates a new JobQueue with the given options.
func NewJobQueue(opts *JobQueueOptions) (*JobQueue, error) {
	if opts.Queue == nil {
		return nil, errors.New("Queue must be provided")
	}
	return &JobQueue{queue: opts.Queue}, nil
}

// EnqueueJob creates a new Job from the given Task and adds it to the queue.
func (q *JobQueue) EnqueueJob(opts *JobOptions) (*Future, error) {
	job, err := NewJob(opts)
	if err != nil {
		return nil, err
	}
	q.queue.Enqueue(job, opts.QueueOptions)
	return job.Future, nil
}

// DequeueJob removes a Job from the queue.
func (q *JobQueue) DequeueJob() (*Job, error) {
	return q.queue.Dequeue()
}

// Close closes the queue.
func (q *JobQueue) Close() {
	q.queue.Close()
}

// Len returns the number of Jobs in the queue.
func (q *JobQueue) Len() int {
	return q.queue.Len()
}
