package jobq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFIFOQueue_EnqueueDequeue(t *testing.T) {
	q := NewFIFOQueue(1)

	// Create a dummy task and job
	task := &DummyTask{}
	job, _ := NewJob(&JobOptions{Task: task})

	// Test Enqueue
	q.Enqueue(job, nil)

	// Test Dequeue
	dequeuedJob, err := q.Dequeue()
	assert.Nil(t, err, "Dequeue should not return an error")
	assert.Equal(t, job, dequeuedJob, "Dequeued job should be the same as the enqueued job")

	// Close the queue and test Dequeue on a closed queue
	q.Close()
	_, err = q.Dequeue()
	assert.Equal(t, ErrQueueClosed, err, "Dequeue should return ErrQueueClosed after queue is closed")
}

func TestFIFOQueue_Len(t *testing.T) {
	q := NewFIFOQueue(2)

	// Length should initially be 0
	assert.Equal(t, 0, q.Len(), "Initial queue length should be 0")

	// Add jobs to the queue
	task := &DummyTask{}

	job1, _ := NewJob(&JobOptions{Task: task})
	job2, _ := NewJob(&JobOptions{Task: task})
	q.Enqueue(job1, nil)
	q.Enqueue(job2, nil)

	// Length should now be 2
	assert.Equal(t, 2, q.Len(), "JobQueue length should be 2 after enqueuing two jobs")
}

func TestFIFOQueue_Close(t *testing.T) {
	q := NewFIFOQueue(1)

	// Close the queue
	q.Close()

	// Try to dequeue after closing
	_, err := q.Dequeue()
	assert.Equal(t, ErrQueueClosed, err, "Dequeue should return ErrQueueClosed after the queue is closed")
}

func TestLIFOQueue_EnqueueDequeue(t *testing.T) {
	q := NewLIFOQueue(2)

	// Create a dummy task and job
	task := &DummyTask{}

	jobOne, _ := NewJob(&JobOptions{Task: task})
	jobTwo, _ := NewJob(&JobOptions{Task: task})

	// Test Enqueue
	q.Enqueue(jobOne, nil)
	q.Enqueue(jobTwo, nil)

	// Test Dequeue
	dequeuedJob, err := q.Dequeue()
	assert.Nil(t, err, "Dequeue should not return an error")
	assert.Equal(t, jobTwo, dequeuedJob, "Dequeued job should be the same as the enqueued job")

	dequeuedJob, err = q.Dequeue()
	assert.Nil(t, err, "Dequeue should not return an error")
	assert.Equal(t, jobOne, dequeuedJob, "Dequeued job should be the same as the enqueued job")
}

func TestLIFOQueue_Len(t *testing.T) {
	q := NewLIFOQueue(2)

	// Length should initially be 0
	assert.Equal(t, 0, q.Len(), "Initial queue length should be 0")

	// Add jobs to the queue
	task := &DummyTask{}

	job1, _ := NewJob(&JobOptions{Task: task})
	job2, _ := NewJob(&JobOptions{Task: task})
	q.Enqueue(job1, nil)
	q.Enqueue(job2, nil)

	// Length should now be 2
	assert.Equal(t, 2, q.Len(), "JobQueue length should be 2 after enqueuing two jobs")
}

func TestLIFOQueue_Close(t *testing.T) {
	q := NewLIFOQueue(1)

	// Close the queue
	q.Close()

	// Try to dequeue after closing
	_, err := q.Dequeue()
	assert.Equal(t, ErrQueueClosed, err, "Dequeue should return ErrQueueClosed after the queue is closed")
}

func TestPriorityQueue_EnqueueDequeue(t *testing.T) {
	q := NewPriorityQueue(10)

	taskHigh := &DummyTask{} // assume DummyTask is defined elsewhere
	taskLow := &DummyTask{}

	jobHigh, _ := NewJob(&JobOptions{Task: taskHigh})
	jobLow, _ := NewJob(&JobOptions{Task: taskLow})

	// Enqueue jobs with different priorities
	q.Enqueue(jobHigh, &PriorityQueueOptions{Priority: 1}) // Higher priority
	q.Enqueue(jobLow, &PriorityQueueOptions{Priority: 10}) // Lower priority

	// Dequeue jobs and check order
	dequeuedJobHigh, err := q.Dequeue()
	assert.Nil(t, err, "Dequeue should not return an error for the first job")
	assert.Equal(t, jobHigh, dequeuedJobHigh, "The first dequeued job should be the one with higher priority")

	dequeuedJobLow, err := q.Dequeue()
	assert.Nil(t, err, "Dequeue should not return an error for the second job")
	assert.Equal(t, jobLow, dequeuedJobLow, "The second dequeued job should be the one with lower priority")
}

func TestPriorityQueue_Len(t *testing.T) {
	q := NewPriorityQueue(10)

	task := &DummyTask{}
	job, _ := NewJob(&JobOptions{Task: task})

	// Check initial length
	assert.Equal(t, 0, q.Len(), "Initial queue length should be 0")

	// Enqueue a job and check length
	q.Enqueue(job, &PriorityQueueOptions{Priority: 5})
	assert.Equal(t, 1, q.Len(), "Queue length should be 1 after enqueuing a job")
}

func TestPriorityQueue_Close(t *testing.T) {
	q := NewPriorityQueue(10)

	// Close the queue
	q.Close()

	// Try to dequeue after closing
	_, err := q.Dequeue()
	assert.Equal(t, ErrQueueClosed, err, "Dequeue should return ErrQueueClosed after the queue is closed")
}
