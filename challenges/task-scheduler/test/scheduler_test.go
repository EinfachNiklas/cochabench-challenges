package taskscheduler

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestNewScheduler_ValidConfig(t *testing.T) {
	config := SchedulerConfig{
		MaxWorkers:      4,
		QueueSize:       100,
		ShutdownTimeout: 5 * time.Second,
		DefaultRetry: RetryPolicy{
			MaxRetries: 3,
			BaseDelay:  time.Millisecond,
			MaxDelay:   time.Second,
			Multiplier: 2.0,
		},
	}

	scheduler, err := NewScheduler(config)
	if err != nil {
		t.Fatalf("NewScheduler() error: %v", err)
	}
	if scheduler == nil {
		t.Fatal("NewScheduler() returned nil")
	}
}

func TestNewScheduler_InvalidConfig(t *testing.T) {
	tests := []struct {
		name   string
		config SchedulerConfig
	}{
		{
			name: "zero MaxWorkers",
			config: SchedulerConfig{
				MaxWorkers:      0,
				QueueSize:       10,
				ShutdownTimeout: time.Second,
			},
		},
		{
			name: "zero QueueSize",
			config: SchedulerConfig{
				MaxWorkers:      4,
				QueueSize:       0,
				ShutdownTimeout: time.Second,
			},
		},
		{
			name: "zero ShutdownTimeout",
			config: SchedulerConfig{
				MaxWorkers:      4,
				QueueSize:       10,
				ShutdownTimeout: 0,
			},
		},
		{
			name: "invalid Multiplier with retries",
			config: SchedulerConfig{
				MaxWorkers:      4,
				QueueSize:       10,
				ShutdownTimeout: time.Second,
				DefaultRetry: RetryPolicy{
					MaxRetries: 3,
					Multiplier: 0,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewScheduler(tt.config)
			if err == nil {
				t.Error("Expected error for invalid config")
			}
		})
	}
}

func TestScheduler_RegisterHandler(t *testing.T) {
	scheduler, _ := NewScheduler(SchedulerConfig{
		MaxWorkers:      2,
		QueueSize:       10,
		ShutdownTimeout: time.Second,
	})

	called := false
	scheduler.RegisterHandler("test", func(task *Task) error {
		called = true
		return nil
	})

	// Starten und Task einreichen
	ctx := context.Background()
	scheduler.Start(ctx)

	task := &Task{
		ID:      "task-1",
		Handler: "test",
	}
	err := scheduler.Submit(task)
	if err != nil {
		t.Fatalf("Submit() error: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	if !called {
		t.Error("Handler was not called")
	}

	shutdownCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	scheduler.Shutdown(shutdownCtx)
}

func TestScheduler_Submit_EmptyID(t *testing.T) {
	scheduler, _ := NewScheduler(SchedulerConfig{
		MaxWorkers:      2,
		QueueSize:       10,
		ShutdownTimeout: time.Second,
	})
	scheduler.RegisterHandler("test", func(task *Task) error { return nil })

	ctx := context.Background()
	scheduler.Start(ctx)

	err := scheduler.Submit(&Task{ID: "", Handler: "test"})
	if err == nil {
		t.Error("Expected error for empty task ID")
	}

	shutdownCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	scheduler.Shutdown(shutdownCtx)
}

func TestScheduler_Submit_UnregisteredHandler(t *testing.T) {
	scheduler, _ := NewScheduler(SchedulerConfig{
		MaxWorkers:      2,
		QueueSize:       10,
		ShutdownTimeout: time.Second,
	})

	ctx := context.Background()
	scheduler.Start(ctx)

	err := scheduler.Submit(&Task{ID: "task-1", Handler: "nonexistent"})
	if err == nil {
		t.Error("Expected error for unregistered handler")
	}

	shutdownCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	scheduler.Shutdown(shutdownCtx)
}

func TestScheduler_GetTask(t *testing.T) {
	scheduler, _ := NewScheduler(SchedulerConfig{
		MaxWorkers:      2,
		QueueSize:       10,
		ShutdownTimeout: time.Second,
	})
	scheduler.RegisterHandler("test", func(task *Task) error { return nil })

	ctx := context.Background()
	scheduler.Start(ctx)

	scheduler.Submit(&Task{ID: "task-1", Handler: "test", Name: "My Task"})

	task, err := scheduler.GetTask("task-1")
	if err != nil {
		t.Fatalf("GetTask() error: %v", err)
	}
	if task.ID != "task-1" {
		t.Errorf("Expected ID 'task-1', got '%s'", task.ID)
	}
	if task.Name != "My Task" {
		t.Errorf("Expected Name 'My Task', got '%s'", task.Name)
	}

	shutdownCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	scheduler.Shutdown(shutdownCtx)
}

func TestScheduler_GetTask_NotFound(t *testing.T) {
	scheduler, _ := NewScheduler(SchedulerConfig{
		MaxWorkers:      2,
		QueueSize:       10,
		ShutdownTimeout: time.Second,
	})

	ctx := context.Background()
	scheduler.Start(ctx)

	_, err := scheduler.GetTask("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent task")
	}

	shutdownCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	scheduler.Shutdown(shutdownCtx)
}

func TestScheduler_Cancel_PendingTask(t *testing.T) {
	scheduler, _ := NewScheduler(SchedulerConfig{
		MaxWorkers:      1,
		QueueSize:       10,
		ShutdownTimeout: time.Second,
	})

	// Handler der blockiert
	blocker := make(chan struct{})
	scheduler.RegisterHandler("blocking", func(task *Task) error {
		<-blocker
		return nil
	})
	scheduler.RegisterHandler("test", func(task *Task) error {
		return nil
	})

	ctx := context.Background()
	scheduler.Start(ctx)

	// Blockierenden Task einreichen
	scheduler.Submit(&Task{ID: "blocking-task", Handler: "blocking"})
	time.Sleep(50 * time.Millisecond)

	// Zweiten Task einreichen (wird in Queue warten)
	scheduler.Submit(&Task{ID: "cancel-me", Handler: "test"})
	time.Sleep(50 * time.Millisecond)

	// Zweiten Task canceln
	err := scheduler.Cancel("cancel-me")
	if err != nil {
		t.Fatalf("Cancel() error: %v", err)
	}

	task, _ := scheduler.GetTask("cancel-me")
	if task != nil && task.Status != StatusCancelled {
		t.Errorf("Expected StatusCancelled, got %v", task.Status)
	}

	close(blocker)
	shutdownCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	scheduler.Shutdown(shutdownCtx)
}

func TestScheduler_TaskCompletion(t *testing.T) {
	scheduler, _ := NewScheduler(SchedulerConfig{
		MaxWorkers:      2,
		QueueSize:       10,
		ShutdownTimeout: time.Second,
	})

	scheduler.RegisterHandler("process", func(task *Task) error {
		task.Result = "done"
		return nil
	})

	ctx := context.Background()
	scheduler.Start(ctx)

	scheduler.Submit(&Task{ID: "task-1", Handler: "process"})

	// Warte auf Fertigstellung
	time.Sleep(200 * time.Millisecond)

	task, err := scheduler.GetTask("task-1")
	if err != nil {
		t.Fatalf("GetTask() error: %v", err)
	}
	if task.Status != StatusCompleted {
		t.Errorf("Expected StatusCompleted, got %v", task.Status)
	}

	shutdownCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	scheduler.Shutdown(shutdownCtx)
}

func TestScheduler_TaskFailure(t *testing.T) {
	scheduler, _ := NewScheduler(SchedulerConfig{
		MaxWorkers:      2,
		QueueSize:       10,
		ShutdownTimeout: time.Second,
	})

	scheduler.RegisterHandler("failing", func(task *Task) error {
		return &NonRetryableError{Err: errors.New("permanent failure")}
	})

	ctx := context.Background()
	scheduler.Start(ctx)

	scheduler.Submit(&Task{ID: "task-1", Handler: "failing"})

	time.Sleep(200 * time.Millisecond)

	task, _ := scheduler.GetTask("task-1")
	if task == nil {
		t.Fatal("GetTask returned nil")
	}
	if task.Status != StatusFailed {
		t.Errorf("Expected StatusFailed, got %v", task.Status)
	}

	shutdownCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	scheduler.Shutdown(shutdownCtx)
}

func TestScheduler_Metrics(t *testing.T) {
	scheduler, _ := NewScheduler(SchedulerConfig{
		MaxWorkers:      2,
		QueueSize:       10,
		ShutdownTimeout: time.Second,
	})

	scheduler.RegisterHandler("ok", func(task *Task) error { return nil })
	scheduler.RegisterHandler("fail", func(task *Task) error {
		return &NonRetryableError{Err: errors.New("fail")}
	})

	ctx := context.Background()
	scheduler.Start(ctx)

	scheduler.Submit(&Task{ID: "ok-1", Handler: "ok", Priority: PriorityHigh})
	scheduler.Submit(&Task{ID: "ok-2", Handler: "ok", Priority: PriorityHigh})
	scheduler.Submit(&Task{ID: "fail-1", Handler: "fail", Priority: PriorityLow})

	time.Sleep(500 * time.Millisecond)

	metrics := scheduler.Metrics()
	if metrics.TotalSubmitted != 3 {
		t.Errorf("Expected TotalSubmitted = 3, got %d", metrics.TotalSubmitted)
	}
	if metrics.TotalCompleted != 2 {
		t.Errorf("Expected TotalCompleted = 2, got %d", metrics.TotalCompleted)
	}
	if metrics.TotalFailed != 1 {
		t.Errorf("Expected TotalFailed = 1, got %d", metrics.TotalFailed)
	}

	shutdownCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	scheduler.Shutdown(shutdownCtx)
}

func TestScheduler_Shutdown(t *testing.T) {
	scheduler, _ := NewScheduler(SchedulerConfig{
		MaxWorkers:      2,
		QueueSize:       10,
		ShutdownTimeout: 2 * time.Second,
	})

	scheduler.RegisterHandler("test", func(task *Task) error {
		time.Sleep(50 * time.Millisecond)
		return nil
	})

	ctx := context.Background()
	scheduler.Start(ctx)

	for i := 0; i < 5; i++ {
		scheduler.Submit(&Task{
			ID:      fmt.Sprintf("task-%d", i),
			Handler: "test",
		})
	}

	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := scheduler.Shutdown(shutdownCtx)
	if err != nil {
		t.Errorf("Shutdown() error: %v", err)
	}
}

func TestScheduler_PriorityOrder(t *testing.T) {
	scheduler, _ := NewScheduler(SchedulerConfig{
		MaxWorkers:      1, // Ein Worker = sequentielle Verarbeitung
		QueueSize:       10,
		ShutdownTimeout: time.Second,
	})

	order := make([]string, 0)
	orderMu := sync.Mutex{}

	// Blockierenden Handler um Queue aufzufüllen
	blocker := make(chan struct{})
	scheduler.RegisterHandler("block", func(task *Task) error {
		<-blocker
		return nil
	})
	scheduler.RegisterHandler("record", func(task *Task) error {
		orderMu.Lock()
		order = append(order, task.ID)
		orderMu.Unlock()
		return nil
	})

	ctx := context.Background()
	scheduler.Start(ctx)

	// Blockierenden Task einreichen
	scheduler.Submit(&Task{ID: "blocker", Handler: "block"})
	time.Sleep(50 * time.Millisecond)

	// Tasks mit verschiedenen Prioritäten einreichen
	scheduler.Submit(&Task{ID: "low", Handler: "record", Priority: PriorityLow})
	scheduler.Submit(&Task{ID: "critical", Handler: "record", Priority: PriorityCritical})
	scheduler.Submit(&Task{ID: "medium", Handler: "record", Priority: PriorityMedium})

	time.Sleep(50 * time.Millisecond)

	// Blocker freigeben
	close(blocker)

	time.Sleep(500 * time.Millisecond)

	orderMu.Lock()
	defer orderMu.Unlock()

	if len(order) < 3 {
		t.Fatalf("Expected 3 recorded tasks, got %d", len(order))
	}

	// Critical sollte zuerst kommen
	if order[0] != "critical" {
		t.Errorf("Expected 'critical' first, got '%s'", order[0])
	}

	shutdownCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	scheduler.Shutdown(shutdownCtx)
}
