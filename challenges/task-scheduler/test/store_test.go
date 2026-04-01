package taskscheduler

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestNewTaskStore(t *testing.T) {
	store := NewTaskStore()
	if store == nil {
		t.Fatal("NewTaskStore() returned nil")
	}
	if store.Count() != 0 {
		t.Errorf("Expected empty store, got Count() = %d", store.Count())
	}
}

func TestTaskStore_Add(t *testing.T) {
	store := NewTaskStore()

	task := &Task{
		ID:        "task-1",
		Name:      "Test Task",
		Priority:  PriorityMedium,
		Status:    StatusPending,
		CreatedAt: time.Now(),
	}

	err := store.Add(task)
	if err != nil {
		t.Fatalf("Add() error: %v", err)
	}
	if store.Count() != 1 {
		t.Errorf("Expected Count() = 1, got %d", store.Count())
	}
}

func TestTaskStore_Add_EmptyID(t *testing.T) {
	store := NewTaskStore()

	task := &Task{ID: "", Name: "No ID"}
	err := store.Add(task)
	if err != ErrTaskIDEmpty {
		t.Errorf("Expected ErrTaskIDEmpty, got %v", err)
	}
}

func TestTaskStore_Add_Duplicate(t *testing.T) {
	store := NewTaskStore()

	task := &Task{ID: "task-1", Name: "First"}
	store.Add(task)

	task2 := &Task{ID: "task-1", Name: "Duplicate"}
	err := store.Add(task2)
	if err != ErrTaskAlreadyExists {
		t.Errorf("Expected ErrTaskAlreadyExists, got %v", err)
	}
}

func TestTaskStore_Get(t *testing.T) {
	store := NewTaskStore()

	original := &Task{
		ID:       "task-1",
		Name:     "Test",
		Priority: PriorityHigh,
		Status:   StatusPending,
	}
	store.Add(original)

	retrieved, err := store.Get("task-1")
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if retrieved.ID != "task-1" {
		t.Errorf("Expected ID 'task-1', got '%s'", retrieved.ID)
	}
	if retrieved.Name != "Test" {
		t.Errorf("Expected Name 'Test', got '%s'", retrieved.Name)
	}
}

func TestTaskStore_Get_ReturnsCopy(t *testing.T) {
	store := NewTaskStore()

	original := &Task{ID: "task-1", Name: "Original", Status: StatusPending}
	store.Add(original)

	// Retrieve copy and modify it
	retrieved, _ := store.Get("task-1")
	retrieved.Name = "Modified"

	// Original in the store should be unchanged
	retrieved2, _ := store.Get("task-1")
	if retrieved2.Name != "Original" {
		t.Errorf("Store should return copies, but internal data was modified: got Name='%s'", retrieved2.Name)
	}
}

func TestTaskStore_Get_NotFound(t *testing.T) {
	store := NewTaskStore()

	_, err := store.Get("nonexistent")
	if err != ErrTaskNotFound {
		t.Errorf("Expected ErrTaskNotFound, got %v", err)
	}
}

func TestTaskStore_Update(t *testing.T) {
	store := NewTaskStore()

	task := &Task{ID: "task-1", Name: "Original", Status: StatusPending}
	store.Add(task)

	updated := &Task{ID: "task-1", Name: "Updated", Status: StatusRunning}
	err := store.Update(updated)
	if err != nil {
		t.Fatalf("Update() error: %v", err)
	}

	retrieved, _ := store.Get("task-1")
	if retrieved.Name != "Updated" {
		t.Errorf("Expected Name 'Updated', got '%s'", retrieved.Name)
	}
	if retrieved.Status != StatusRunning {
		t.Errorf("Expected StatusRunning, got %v", retrieved.Status)
	}
}

func TestTaskStore_Update_NotFound(t *testing.T) {
	store := NewTaskStore()

	task := &Task{ID: "nonexistent", Name: "Ghost"}
	err := store.Update(task)
	if err != ErrTaskNotFound {
		t.Errorf("Expected ErrTaskNotFound, got %v", err)
	}
}

func TestTaskStore_Delete(t *testing.T) {
	store := NewTaskStore()

	task := &Task{ID: "task-1", Name: "To Delete"}
	store.Add(task)

	err := store.Delete("task-1")
	if err != nil {
		t.Fatalf("Delete() error: %v", err)
	}
	if store.Count() != 0 {
		t.Errorf("Expected Count() = 0 after delete, got %d", store.Count())
	}

	_, err = store.Get("task-1")
	if err != ErrTaskNotFound {
		t.Error("Task should not be found after deletion")
	}
}

func TestTaskStore_Delete_NotFound(t *testing.T) {
	store := NewTaskStore()

	err := store.Delete("nonexistent")
	if err != ErrTaskNotFound {
		t.Errorf("Expected ErrTaskNotFound, got %v", err)
	}
}

func TestTaskStore_List_NoFilter(t *testing.T) {
	store := NewTaskStore()

	store.Add(&Task{ID: "t1", Status: StatusPending, Priority: PriorityLow})
	store.Add(&Task{ID: "t2", Status: StatusRunning, Priority: PriorityHigh})
	store.Add(&Task{ID: "t3", Status: StatusCompleted, Priority: PriorityMedium})

	results := store.List(TaskFilter{})
	if len(results) != 3 {
		t.Errorf("Expected 3 results with no filter, got %d", len(results))
	}
}

func TestTaskStore_List_FilterByStatus(t *testing.T) {
	store := NewTaskStore()

	store.Add(&Task{ID: "t1", Status: StatusPending})
	store.Add(&Task{ID: "t2", Status: StatusRunning})
	store.Add(&Task{ID: "t3", Status: StatusPending})

	status := StatusPending
	results := store.List(TaskFilter{Status: &status})
	if len(results) != 2 {
		t.Errorf("Expected 2 pending tasks, got %d", len(results))
	}
	for _, task := range results {
		if task.Status != StatusPending {
			t.Errorf("Expected StatusPending, got %v", task.Status)
		}
	}
}

func TestTaskStore_List_FilterByPriority(t *testing.T) {
	store := NewTaskStore()

	store.Add(&Task{ID: "t1", Priority: PriorityHigh})
	store.Add(&Task{ID: "t2", Priority: PriorityLow})
	store.Add(&Task{ID: "t3", Priority: PriorityHigh})

	priority := PriorityHigh
	results := store.List(TaskFilter{Priority: &priority})
	if len(results) != 2 {
		t.Errorf("Expected 2 high-priority tasks, got %d", len(results))
	}
}

func TestTaskStore_List_FilterByHandler(t *testing.T) {
	store := NewTaskStore()

	store.Add(&Task{ID: "t1", Handler: "email"})
	store.Add(&Task{ID: "t2", Handler: "sms"})
	store.Add(&Task{ID: "t3", Handler: "email"})

	results := store.List(TaskFilter{Handler: "email"})
	if len(results) != 2 {
		t.Errorf("Expected 2 email tasks, got %d", len(results))
	}
}

func TestTaskStore_List_CombinedFilter(t *testing.T) {
	store := NewTaskStore()

	store.Add(&Task{ID: "t1", Status: StatusPending, Priority: PriorityHigh, Handler: "email"})
	store.Add(&Task{ID: "t2", Status: StatusPending, Priority: PriorityLow, Handler: "email"})
	store.Add(&Task{ID: "t3", Status: StatusRunning, Priority: PriorityHigh, Handler: "email"})
	store.Add(&Task{ID: "t4", Status: StatusPending, Priority: PriorityHigh, Handler: "sms"})

	status := StatusPending
	priority := PriorityHigh
	results := store.List(TaskFilter{
		Status:   &status,
		Priority: &priority,
		Handler:  "email",
	})
	if len(results) != 1 {
		t.Errorf("Expected 1 task matching all filters, got %d", len(results))
	}
	if len(results) > 0 && results[0].ID != "t1" {
		t.Errorf("Expected 't1', got '%s'", results[0].ID)
	}
}

func TestTaskStore_ConcurrentAccess(t *testing.T) {
	store := NewTaskStore()
	var wg sync.WaitGroup

	// Concurrent schreiben
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			task := &Task{
				ID:       fmt.Sprintf("task-%d", id),
				Name:     fmt.Sprintf("Task %d", id),
				Status:   StatusPending,
				Priority: Priority(id % 4),
			}
			store.Add(task)
		}(i)
	}
	wg.Wait()

	if store.Count() != 100 {
		t.Errorf("Expected 100 tasks after concurrent writes, got %d", store.Count())
	}

	// Concurrent lesen und schreiben
	for i := 0; i < 100; i++ {
		wg.Add(2)
		go func(id int) {
			defer wg.Done()
			store.Get(fmt.Sprintf("task-%d", id))
		}(i)
		go func(id int) {
			defer wg.Done()
			task := &Task{
				ID:     fmt.Sprintf("task-%d", id),
				Name:   fmt.Sprintf("Updated %d", id),
				Status: StatusCompleted,
			}
			store.Update(task)
		}(i)
	}
	wg.Wait()
}
