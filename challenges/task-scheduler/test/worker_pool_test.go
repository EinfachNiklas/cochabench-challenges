package taskscheduler

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewWorkerPool(t *testing.T) {
	pool := NewWorkerPool(4)
	if pool == nil {
		t.Fatal("NewWorkerPool() returned nil")
	}
}

func TestWorkerPool_StartAndShutdown(t *testing.T) {
	pool := NewWorkerPool(4)
	ctx := context.Background()

	pool.Start(ctx)

	shutdownCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	err := pool.Shutdown(shutdownCtx)
	if err != nil {
		t.Fatalf("Shutdown() error: %v", err)
	}
}

func TestWorkerPool_SubmitBeforeStart(t *testing.T) {
	pool := NewWorkerPool(2)

	task := &Task{ID: "task-1", Status: StatusPending}
	handler := func(task *Task) error { return nil }

	err := pool.Submit(task, handler)
	if err != ErrPoolNotRunning {
		t.Errorf("Expected ErrPoolNotRunning, got %v", err)
	}
}

func TestWorkerPool_SubmitAndProcess(t *testing.T) {
	pool := NewWorkerPool(2)
	ctx := context.Background()
	pool.Start(ctx)

	var processed int64

	handler := func(task *Task) error {
		atomic.AddInt64(&processed, 1)
		return nil
	}

	task := &Task{ID: "task-1", Status: StatusPending}
	err := pool.Submit(task, handler)
	if err != nil {
		t.Fatalf("Submit() error: %v", err)
	}

	// Wait for processing

	if atomic.LoadInt64(&processed) != 1 {
		t.Errorf("Expected 1 processed task, got %d", atomic.LoadInt64(&processed))
	}

	shutdownCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	pool.Shutdown(shutdownCtx)
}

func TestWorkerPool_SetsTaskStatus(t *testing.T) {
	pool := NewWorkerPool(1)
	ctx := context.Background()
	pool.Start(ctx)

	done := make(chan struct{})

	task := &Task{ID: "task-1", Status: StatusPending}
	handler := func(task *Task) error {
		defer close(done)
		return nil
	}

	pool.Submit(task, handler)

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for task completion")
	}

	// Kurz warten damit der Worker den Status setzen kann
	time.Sleep(50 * time.Millisecond)

	if task.Status != StatusCompleted {
		t.Errorf("Expected StatusCompleted, got %v", task.Status)
	}
	if task.StartedAt.IsZero() {
		t.Error("Expected StartedAt to be set")
	}
	if task.CompletedAt.IsZero() {
		t.Error("Expected CompletedAt to be set")
	}
}

func TestWorkerPool_HandlerError_SetsStatusFailed(t *testing.T) {
	pool := NewWorkerPool(1)
	ctx := context.Background()
	pool.Start(ctx)

	done := make(chan struct{})

	task := &Task{ID: "task-1", Status: StatusPending}
	handler := func(task *Task) error {
		defer close(done)
		return fmt.Errorf("something went wrong")
	}

	pool.Submit(task, handler)

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for task")
	}

	time.Sleep(50 * time.Millisecond)

	if task.Status != StatusFailed {
		t.Errorf("Expected StatusFailed after error, got %v", task.Status)
	}
	if task.Error == nil {
		t.Error("Expected Error to be set")
	}
}

func TestWorkerPool_MultipleTasks(t *testing.T) {
	pool := NewWorkerPool(4)
	ctx := context.Background()
	pool.Start(ctx)

	var processed int64
	taskCount := 20

	handler := func(task *Task) error {
		time.Sleep(10 * time.Millisecond)
		atomic.AddInt64(&processed, 1)
		return nil
	}

	for i := 0; i < taskCount; i++ {
		task := &Task{
			ID:     fmt.Sprintf("task-%d", i),
			Status: StatusPending,
		}
		err := pool.Submit(task, handler)
		if err != nil {
			t.Fatalf("Submit() error for task-%d: %v", i, err)
		}
	}

	// Wait for processing

	count := atomic.LoadInt64(&processed)
	if count != int64(taskCount) {
		t.Errorf("Expected %d processed tasks, got %d", taskCount, count)
	}

	shutdownCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	pool.Shutdown(shutdownCtx)
}

func TestWorkerPool_ProcessedCount(t *testing.T) {
	pool := NewWorkerPool(2)
	ctx := context.Background()
	pool.Start(ctx)

	handler := func(task *Task) error { return nil }

	for i := 0; i < 5; i++ {
		pool.Submit(&Task{ID: fmt.Sprintf("task-%d", i), Status: StatusPending}, handler)
	}

	time.Sleep(200 * time.Millisecond)

	if pool.ProcessedCount() != 5 {
		t.Errorf("Expected ProcessedCount() = 5, got %d", pool.ProcessedCount())
	}

	shutdownCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	pool.Shutdown(shutdownCtx)
}

func TestWorkerPool_GracefulShutdown(t *testing.T) {
	pool := NewWorkerPool(1)
	ctx := context.Background()
	pool.Start(ctx)

	started := make(chan struct{})
	handler := func(task *Task) error {
		close(started)
		time.Sleep(200 * time.Millisecond)
		return nil
	}

	pool.Submit(&Task{ID: "long-task", Status: StatusPending}, handler)

	// Warte bis der Task gestartet ist
	<-started

	// Shutdown sollte auf den laufenden Task warten
	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := pool.Shutdown(shutdownCtx)
	if err != nil {
		t.Errorf("Shutdown() should succeed, got error: %v", err)
	}

	if pool.ProcessedCount() != 1 {
		t.Errorf("Expected 1 processed task after shutdown, got %d", pool.ProcessedCount())
	}
}

func TestWorkerPool_ConcurrentSubmit(t *testing.T) {
	pool := NewWorkerPool(8)
	ctx := context.Background()
	pool.Start(ctx)

	var processed int64
	handler := func(task *Task) error {
		atomic.AddInt64(&processed, 1)
		return nil
	}

	// Concurrent submit von vielen Tasks
	done := make(chan struct{})
	go func() {
		for i := 0; i < 50; i++ {
			go func(id int) {
				pool.Submit(&Task{
					ID:     fmt.Sprintf("task-%d", id),
					Status: StatusPending,
				}, handler)
			}(i)
		}
		close(done)
	}()
	<-done

	time.Sleep(500 * time.Millisecond)

	shutdownCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	pool.Shutdown(shutdownCtx)

	count := atomic.LoadInt64(&processed)
	t.Logf("Processed %d tasks concurrently", count)
	if count == 0 {
		t.Error("Expected at least some tasks to be processed")
	}
}
