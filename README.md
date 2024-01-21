# JobQ


## Overview

JobQ is a lightweight, in-memory job queue implementation in Go designed for asynchronous task processing within the same process. It allows for simple job scheduling and processing, suitable for applications that require asynchronous task handling without the need for distributed messaging systems.

## Features

- **Simple API**: Easy-to-use functions for creating jobs and processing them.
- **Concurrency Safe**: Safely handles multiple concurrent job enqueuers and workers.
- **Retry Logic**: Supports retry logic for jobs.
- **Future Results**: Implements a Future pattern for job results, allowing asynchronous result retrieval.

## Installation

```bash
go get gitlab.com/emergentmethods/jobq
```

## Usage

### Basic Concepts

- **Job**: A unit of work that needs to be executed.
- **Task**: An interface that your work units must implement.
- **Future**: A mechanism to retrieve the result of a job asynchronously.
- **JobQueue**: A queue that holds and manages the jobs.
- **WorkerPool**: A pool of workers that execute jobs from the queue.

### Creating a Task

Implement the `Task`` interface for the work you want to perform:

```go
type MyTask struct {
    // Task-specific fields
}

func (t *MyTask) Execute(ctx context.Context) (interface{}, error) {
    // Task logic
    return result, nil
}
```

### Creating a JobQueue and enqueueing a Job

```go
func main() {
    // We create a JobQueue with a FIFO queue of size 10
    queue, err := jobq.NewJobQueue(&jobq.JobQueueOptions{
        Queue: jobq.NewFIFOQueue(10),
    })
    if err != nil {
        panic(err)
    }
    // We create a WorkerPool with 10 workers
    pool := jobq.NewWorkerPool(queue, 10)
    // Closing the pool will stop all workers and close the queue
    defer pool.Close()

    // We create a task and enqueue it
    // We pass a context, a unique Job ID, and a the maximum number of retries. Using 0
    // for the maximum amount of retries means the task will only be attempted once.
    // This allows us to cancel the task via the context and specify how many times it 
    // should be retried if it fails.
    future, _ := queue.EnqueueJob(&JobOptions{
        Task: &MyTask{},
        Ctx: context.WithTimeout(3*time.Second),
        MaxRetries: 2,
        ID: uuid.New(),
    })

    // Optionally, wait for the result. This is a blocking operation,
    // and will wait until the job is processed.
    result, _ := future.Result()
    fmt.Println("Job result:", result)
}
```

### Queue implementations

JobQ provides two queue implementations:

- **FIFOQueue**: A FIFO queue that holds jobs in the order they were enqueued.
- **LIFOQueue**: A LIFO queue that holds jobs in the reverse order they were enqueued.
- **PriorityQueue**: A priority queue that holds jobs in priority order. Jobs with a lower priority value will be processed first.

#### FIFOQueue

```go
queue, err := jobq.NewJobQueue(&jobq.JobQueueOptions{
    Queue: jobq.NewFIFOQueue(10),
})
```

#### LIFOQueue

```go
queue, err := jobq.NewJobQueue(&jobq.JobQueueOptions{
    Queue: jobq.NewLIFOQueue(10),
})
```

#### PriorityQueue

```go
queue, err := jobq.NewJobQueue(&jobq.JobQueueOptions{
    Queue: jobq.NewPriorityQueue(10),
})
if err != nil {
    panic(err)
}

// We can set the priority of a job when we enqueue it
future, _ := queue.EnqueueJob(&JobOptions{
    Task: &MyTask{},
    Ctx: context.WithTimeout(3*time.Second),
    MaxRetries: 2,
    ID: uuid.New(),
    QueueOptions: &PriorityQueueOptions{
        Priority: 1,
    },
})
```

```