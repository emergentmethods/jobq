package jobq

import (
	"context"
)

type Task interface {
	Execute() (interface{}, error)
}

type Job struct {
	ctx    context.Context
	Task   Task
	Future *Future
}

func NewJob(ctx context.Context, task Task) Job {
	return Job{
		ctx:    ctx,
		Task:   task,
		Future: NewFuture(),
	}
}

func (j Job) Run() {
	done := make(chan struct{}, 1)

	go func() {
		defer close(done)
		result, err := j.Task.Execute()
		j.Future.SetResult(result, err)
	}()

	select {
	case <-j.ctx.Done():
		j.Future.SetResult(nil, j.ctx.Err())
	case <-done:
		return
	}
}

func (j Job) Result() (interface{}, error) {
	return j.Future.Result()
}

func (j Job) Empty() bool {
	return j.Task == nil
}
