package main

import (
	"context"
	"fmt"
	"time"

	"gitlab.com/emergentmethods/jobq"
)

type HelloTask struct{}

func (t *HelloTask) Execute() (interface{}, error) {
	time.Sleep(2 * time.Second)
	return "Hello, People!", nil
}

func main() {
	// Create a new in-memory queue
	q := jobq.NewQueue(10)
	pool := jobq.NewWorkerPool(q, 5)
	defer pool.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	fut, err := q.EnqueueTask(ctx, &HelloTask{})
	if err != nil {
		panic(err)
	}

	result, err := fut.Result()
	if err != nil {
		fmt.Printf("Task Error: %s\n", err)
		return
	}
	println(result.(string))

	// Output:
	// Hello, World
}
