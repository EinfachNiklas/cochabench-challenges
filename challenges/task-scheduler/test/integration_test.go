package taskscheduler

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestIntegration_FullWorkflow testet den kompletten Workflow: Submit → Execute → Complete
func TestIntegration_FullWorkflow(t *testing.T) {
	config := SchedulerConfig{
		MaxWorkers:      4,
		QueueSize:       100,
		ShutdownTimeout: 5 * time.Second,
		DefaultRetry: RetryPolicy{
			MaxRetries: 2,
			BaseDelay:  time.Millisecond,
			MaxDelay:   10 * time.Millisecond,
			Multiplier: 2.0,
		},
	}

	scheduler, err := NewScheduler(config)
	if err != nil {
		t.Fatalf("NewScheduler() error: %v", err)
	}

	// Handler registrieren
	scheduler.RegisterHandler("uppercase", func(task *Task) error {
		input, ok := task.Payload["input"].(string)
		if !ok {
			return &NonRetryableError{Err: errors.New("invalid payload")}
		}
		result := ""
		for _, c := range input {
			if c >= 'a' && c <= 'z' {
				result += string(c - 32)
			} else {
				result += string(c)
			}
		}
		task.Result = result
		return nil
	})

	ctx := context.Background()
	scheduler.Start(ctx)

	// Mehrere Tasks einreichen
	tasks := []struct {
		id       string
		priority Priority
		input    string
	}{
		{"task-1", PriorityCritical, "hello"},
		{"task-2", PriorityHigh, "world"},
		{"task-3", PriorityLow, "test"},
		{"task-4", PriorityMedium, "data"},
	}

	for _, tt := range tasks {
		err := scheduler.Submit(&Task{
			ID:       tt.id,
			Handler:  "uppercase",
			Priority: tt.priority,
			Payload:  map[string]interface{}{"input": tt.input},
		})
		if err != nil {
			t.Fatalf("Submit(%s) error: %v", tt.id, err)
		}
	}

	// Warte auf Fertigstellung
	time.Sleep(500 * time.Millisecond)

	// Prüfe Ergebnisse
	expected := map[string]string{
		"task-1": "HELLO",
		"task-2": "WORLD",
		"task-3": "TEST",
		"task-4": "DATA",
	}

	for id, expectedResult := range expected {
		task, err := scheduler.GetTask(id)
		if err != nil {
			t.Errorf("GetTask(%s) error: %v", id, err)
			continue
		}
		if task.Status != StatusCompleted {
			t.Errorf("Task %s: expected StatusCompleted, got %v", id, task.Status)
			continue
		}
		if task.Result != expectedResult {
			t.Errorf("Task %s: expected result '%s', got '%v'", id, expectedResult, task.Result)
		}
	}

	// Prüfe Metrics
	metrics := scheduler.Metrics()
	if metrics.TotalSubmitted != 4 {
		t.Errorf("Expected 4 submitted, got %d", metrics.TotalSubmitted)
	}
	if metrics.TotalCompleted != 4 {
		t.Errorf("Expected 4 completed, got %d", metrics.TotalCompleted)
	}

	shutdownCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	scheduler.Shutdown(shutdownCtx)
}

// TestIntegration_Dependencies testet Task-Abhängigkeiten
func TestIntegration_Dependencies(t *testing.T) {
	config := SchedulerConfig{
		MaxWorkers:      2,
		QueueSize:       100,
		ShutdownTimeout: 5 * time.Second,
	}

	scheduler, _ := NewScheduler(config)

	executionOrder := make([]string, 0)
	var orderMu sync.Mutex

	scheduler.RegisterHandler("track", func(task *Task) error {
		orderMu.Lock()
		executionOrder = append(executionOrder, task.ID)
		orderMu.Unlock()
		time.Sleep(50 * time.Millisecond)
		return nil
	})

	ctx := context.Background()
	scheduler.Start(ctx)

	// Task A: Keine Dependencies
	scheduler.Submit(&Task{
		ID:      "task-a",
		Handler: "track",
	})

	// Task B: Abhängig von A
	scheduler.Submit(&Task{
		ID:           "task-b",
		Handler:      "track",
		Dependencies: []string{"task-a"},
	})

	// Warte auf Fertigstellung
	time.Sleep(time.Second)

	orderMu.Lock()
	defer orderMu.Unlock()

	if len(executionOrder) < 2 {
		t.Fatalf("Expected at least 2 executed tasks, got %d", len(executionOrder))
	}

	// Task A muss vor Task B kommen
	aIndex := -1
	bIndex := -1
	for i, id := range executionOrder {
		if id == "task-a" {
			aIndex = i
		}
		if id == "task-b" {
			bIndex = i
		}
	}

	if aIndex == -1 {
		t.Error("task-a was not executed")
	}
	if bIndex == -1 {
		t.Error("task-b was not executed")
	}
	if aIndex >= bIndex {
		t.Errorf("task-a (index %d) should execute before task-b (index %d)", aIndex, bIndex)
	}

	shutdownCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	scheduler.Shutdown(shutdownCtx)
}

// TestIntegration_RetryWithBackoff testet Retry mit Backoff
func TestIntegration_RetryWithBackoff(t *testing.T) {
	config := SchedulerConfig{
		MaxWorkers:      2,
		QueueSize:       100,
		ShutdownTimeout: 5 * time.Second,
		DefaultRetry: RetryPolicy{
			MaxRetries: 3,
			BaseDelay:  time.Millisecond,
			MaxDelay:   50 * time.Millisecond,
			Multiplier: 2.0,
		},
	}

	scheduler, _ := NewScheduler(config)

	var attempts int64
	scheduler.RegisterHandler("flaky", func(task *Task) error {
		count := atomic.AddInt64(&attempts, 1)
		if count < 3 {
			return &RetryableError{Err: errors.New("temporary error")}
		}
		task.Result = "success after retries"
		return nil
	})

	ctx := context.Background()
	scheduler.Start(ctx)

	scheduler.Submit(&Task{
		ID:         "retry-task",
		Handler:    "flaky",
		MaxRetries: 3,
	})

	// Warte auf Retries
	time.Sleep(time.Second)

	task, _ := scheduler.GetTask("retry-task")
	if task == nil {
		t.Fatal("Task not found")
	}
	if task.Status != StatusCompleted {
		t.Errorf("Expected StatusCompleted after retries, got %v (error: %v)", task.Status, task.Error)
	}
	if task.Result != "success after retries" {
		t.Errorf("Expected result 'success after retries', got '%v'", task.Result)
	}

	metrics := scheduler.Metrics()
	if metrics.TotalRetried == 0 {
		t.Error("Expected some retries to be recorded")
	}

	shutdownCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	scheduler.Shutdown(shutdownCtx)
}

// TestIntegration_CircularDependency testet Erkennung zirkulärer Abhängigkeiten
func TestIntegration_CircularDependency(t *testing.T) {
	config := SchedulerConfig{
		MaxWorkers:      2,
		QueueSize:       100,
		ShutdownTimeout: 5 * time.Second,
	}

	scheduler, _ := NewScheduler(config)
	scheduler.RegisterHandler("test", func(task *Task) error { return nil })

	ctx := context.Background()
	scheduler.Start(ctx)

	// Task A abhängig von Task B
	scheduler.Submit(&Task{ID: "task-a", Handler: "test"})
	scheduler.Submit(&Task{
		ID:           "task-b",
		Handler:      "test",
		Dependencies: []string{"task-a"},
	})

	// Task C abhängig von B, und versuche A abhängig von C zu machen → Zyklus
	scheduler.Submit(&Task{
		ID:           "task-c",
		Handler:      "test",
		Dependencies: []string{"task-b"},
	})

	// Dieser Submit sollte fehlschlagen (würde Zyklus erzeugen)
	err := scheduler.Submit(&Task{
		ID:           "task-d",
		Handler:      "test",
		Dependencies: []string{"task-c", "task-d"}, // Self-dependency = Zyklus
	})
	if err == nil {
		t.Error("Expected error for circular dependency")
	}

	shutdownCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	scheduler.Shutdown(shutdownCtx)
}

// TestIntegration_ManyTasks testet den Scheduler mit vielen Tasks
func TestIntegration_ManyTasks(t *testing.T) {
	config := SchedulerConfig{
		MaxWorkers:      8,
		QueueSize:       500,
		ShutdownTimeout: 10 * time.Second,
	}

	scheduler, _ := NewScheduler(config)

	var completed int64
	scheduler.RegisterHandler("counter", func(task *Task) error {
		atomic.AddInt64(&completed, 1)
		return nil
	})

	ctx := context.Background()
	scheduler.Start(ctx)

	taskCount := 100
	for i := 0; i < taskCount; i++ {
		err := scheduler.Submit(&Task{
			ID:       fmt.Sprintf("task-%d", i),
			Handler:  "counter",
			Priority: Priority(i % 4),
		})
		if err != nil {
			t.Fatalf("Submit(task-%d) error: %v", i, err)
		}
	}

	// Warte auf Fertigstellung
	deadline := time.After(5 * time.Second)
	for {
		if atomic.LoadInt64(&completed) >= int64(taskCount) {
			break
		}
		select {
		case <-deadline:
			t.Fatalf("Timeout: only %d/%d tasks completed", atomic.LoadInt64(&completed), taskCount)
		default:
			time.Sleep(50 * time.Millisecond)
		}
	}

	metrics := scheduler.Metrics()
	if metrics.TotalSubmitted != int64(taskCount) {
		t.Errorf("Expected %d submitted, got %d", taskCount, metrics.TotalSubmitted)
	}
	if metrics.TotalCompleted != int64(taskCount) {
		t.Errorf("Expected %d completed, got %d", taskCount, metrics.TotalCompleted)
	}

	shutdownCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	scheduler.Shutdown(shutdownCtx)
}

// TestIntegration_GracefulShutdown testet dass der Scheduler laufende Tasks abschließt
func TestIntegration_GracefulShutdown(t *testing.T) {
	config := SchedulerConfig{
		MaxWorkers:      2,
		QueueSize:       100,
		ShutdownTimeout: 5 * time.Second,
	}

	scheduler, _ := NewScheduler(config)

	var completed int64
	scheduler.RegisterHandler("slow", func(task *Task) error {
		time.Sleep(100 * time.Millisecond)
		atomic.AddInt64(&completed, 1)
		return nil
	})

	ctx := context.Background()
	scheduler.Start(ctx)

	// Tasks einreichen
	for i := 0; i < 5; i++ {
		scheduler.Submit(&Task{
			ID:      fmt.Sprintf("task-%d", i),
			Handler: "slow",
		})
	}

	// Kurz warten, dann Shutdown
	time.Sleep(50 * time.Millisecond)

	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err := scheduler.Shutdown(shutdownCtx)
	if err != nil {
		t.Errorf("Shutdown() error: %v", err)
	}

	// Nach Shutdown sollten einige Tasks fertig sein
	count := atomic.LoadInt64(&completed)
	if count == 0 {
		t.Error("Expected at least some tasks to complete before shutdown")
	}
	t.Logf("Completed %d tasks before/during shutdown", count)
}
