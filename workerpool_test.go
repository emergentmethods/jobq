package jobq

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWorkerPool_ProcessingJobs(t *testing.T) {
	q, err := NewJobQueue(&JobQueueOptions{
		Queue: NewFIFOQueue(1),
	})
	assert.Nil(t, err, "Expected no error from NewJobQueue")
	pool := NewWorkerPool(q, 2)
	defer pool.Close()

	var wg sync.WaitGroup
	numJobs := 5
	results := make([]string, numJobs)

	// Enqueue and process jobs
	for i := 0; i < numJobs; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			task := &DummyTask{result: "job" + fmt.Sprint(index), delay: 10 * time.Millisecond}
			fut, _ := q.EnqueueJob(&JobOptions{Task: task})
			res, _ := fut.Result()
			results[index] = res.(string)
		}(i)
	}

	// Wait for all jobs to complete
	wg.Wait()

	// Check if all results are as expected
	for i := 0; i < numJobs; i++ {
		expected := "job" + fmt.Sprint(i)
		assert.Contains(t, results, expected, "Expected result %s in results slice", expected)
	}
}

func TestWorkerPool_JobErrorHandling(t *testing.T) {
	q, err := NewJobQueue(&JobQueueOptions{
		Queue: NewFIFOQueue(1),
	})
	assert.Nil(t, err, "Expected no error from NewJobQueue")
	pool := NewWorkerPool(q, 1)
	defer pool.Close()

	task := &DummyTask{err: errors.New("task error")}
	fut, _ := q.EnqueueJob(&JobOptions{Task: task})

	_, err = fut.Result()
	assert.NotNil(t, err, "Expected an error from the task execution")
}

func benchmarkWorkerPool(numJobs int, numWorkers int, b *testing.B) {
	for n := 0; n < b.N; n++ {
		queue, _ := NewJobQueue(&JobQueueOptions{Queue: NewFIFOQueue(100)})
		pool := NewWorkerPool(queue, numWorkers)

		var wg sync.WaitGroup
		wg.Add(numJobs)

		for i := 0; i < numJobs; i++ {
			go func() {
				defer wg.Done()
				task := &DummyTask{result: "result", delay: 3 * time.Millisecond}
				_, _ = queue.EnqueueJob(&JobOptions{Task: task})
			}()
		}

		wg.Wait()
		b.StartTimer()
		pool.Close()
		b.StopTimer()
	}
}

func BenchmarkWorkerPool_NumJobs1NumWorkers1(b *testing.B)   { benchmarkWorkerPool(1, 1, b) }
func BenchmarkWorkerPool_NumJobs10NumWorkers1(b *testing.B)  { benchmarkWorkerPool(10, 1, b) }
func BenchmarkWorkerPool_NumJobs100NumWorkers1(b *testing.B) { benchmarkWorkerPool(100, 1, b) }
func BenchmarkWorkerPool_NumJobs1NumWorkers2(b *testing.B)   { benchmarkWorkerPool(1, 2, b) }
func BenchmarkWorkerPool_NumJobs10NumWorkers2(b *testing.B)  { benchmarkWorkerPool(10, 2, b) }
func BenchmarkWorkerPool_NumJobs100NumWorkers2(b *testing.B) { benchmarkWorkerPool(100, 2, b) }
func BenchmarkWorkerPool_NumJobs1NumWorkers4(b *testing.B)   { benchmarkWorkerPool(1, 4, b) }
func BenchmarkWorkerPool_NumJobs10NumWorkers4(b *testing.B)  { benchmarkWorkerPool(10, 4, b) }
func BenchmarkWorkerPool_NumJobs100NumWorkers4(b *testing.B) { benchmarkWorkerPool(100, 4, b) }
