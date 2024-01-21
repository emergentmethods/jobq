package jobq

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// DummyTask for testing
type DummyTask struct {
	result interface{}
	err    error
	delay  time.Duration // delay before returning result
}

func (dt *DummyTask) Execute(ctx context.Context) (interface{}, error) {
	time.Sleep(dt.delay)
	return dt.result, dt.err
}

func TestJobQueue_NewJobQueue(t *testing.T) {
	q, err := NewJobQueue(&JobQueueOptions{
		Queue: NewFIFOQueue(1),
	})
	assert.Nil(t, err, "Expected no error from NewJobQueue")
	assert.NotNil(t, q, "Expected a non-nil JobQueue")

	q.Close()

	_, err = NewJobQueue(&JobQueueOptions{})
	assert.NotNil(t, err, "Expected an error from NewJobQueue")
}

func TestJobQueue_GetStatus(t *testing.T) {
	task := &DummyTask{}
	job, _ := NewJob(&JobOptions{Task: task})

	assert.Equal(t, StatusPending, job.GetStatus(), "Job should have status StatusRunning")
}

func TestJobQueue_SetStatus(t *testing.T) {
	task := &DummyTask{}
	job, _ := NewJob(&JobOptions{Task: task})

	job.SetStatus(StatusRunning)
	assert.Equal(t, StatusRunning, job.GetStatus(), "Job should have status StatusRunning")
}

func TestJobQueue_Len(t *testing.T) {
	q, err := NewJobQueue(&JobQueueOptions{
		Queue: NewFIFOQueue(1),
	})
	assert.Nil(t, err, "Expected no error from NewJobQueue")

	assert.Equal(t, 0, q.Len(), "JobQueue length should be 0")

	task := &DummyTask{}
	q.EnqueueJob(&JobOptions{Task: task})
	assert.Equal(t, 1, q.Len(), "JobQueue length should be 1 after enqueuing a job")

	q.Close()
}

func TestJob_Empty(t *testing.T) {
	task := &DummyTask{}
	job, _ := NewJob(&JobOptions{Task: task})

	assert.False(t, job.Empty(), "Job should not be empty")

	job.Task = nil
	assert.True(t, job.Empty(), "Job should be empty")
}

func TestJob_Result(t *testing.T) {
	task := &DummyTask{result: "success"}
	job, _ := NewJob(&JobOptions{Task: task})

	job.Run()

	result, err := job.Result()
	assert.Nil(t, err, "Job should not return an error")
	assert.Equal(t, "success", result, "Job should return the task's result")
}

func TestJob_Run(t *testing.T) {
	task := &DummyTask{result: "success", delay: 10 * time.Millisecond}
	job, _ := NewJob(&JobOptions{Task: task})

	go job.Run()

	result, err := job.Future.Result()
	assert.Nil(t, err, "Job should not return an error")
	assert.Equal(t, "success", result, "Job should return the task's result")
}

func TestJob_Run_withRetries(t *testing.T) {
	task := &DummyTask{result: "success", delay: 10 * time.Millisecond}
	job, _ := NewJob(&JobOptions{Task: task})

	go job.Run()

	result, err := job.Future.Result()
	assert.Nil(t, err, "Job should not return an error")
	assert.Equal(t, "success", result, "Job should return the task's result")

	assert.Equal(t, 0, job.retries, "Job should have 1 retries")
	assert.Equal(t, 1, job.maxRetries, "Job should have 1 max retries")

	task = &DummyTask{result: "error", err: errors.New("Failure"), delay: 10 * time.Millisecond}
	job, _ = NewJob(&JobOptions{
		Task:       task,
		MaxRetries: 2,
	})

	go job.Run()

	_, err = job.Future.Result()
	assert.NotNil(t, err, "Job should return an error")
	assert.Equal(t, "Failure", err.Error(), "Job should return the task's error")
	assert.Equal(t, 2, job.retries, "Job should have 2 retries")
	assert.Equal(t, 2, job.maxRetries, "Job should have 2 max retries")
}
