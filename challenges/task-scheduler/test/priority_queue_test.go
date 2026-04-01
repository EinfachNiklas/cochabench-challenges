package taskscheduler

import (
	"fmt"
	"testing"
	"time"
)

func TestNewPriorityQueue(t *testing.T) {
	pq := NewPriorityQueue()
	if pq == nil {
		t.Fatal("NewPriorityQueue() returned nil")
	}
	if pq.Len() != 0 {
		t.Errorf("Expected empty queue, got Len() = %d", pq.Len())
	}
}

func TestPriorityQueue_PushPop_SingleItem(t *testing.T) {
	pq := NewPriorityQueue()

	task := &Task{
		ID:        "task-1",
		Name:      "Test Task",
		Priority:  PriorityMedium,
		CreatedAt: time.Now(),
	}

	pq.Push(task)

	if pq.Len() != 1 {
		t.Errorf("Expected Len() = 1, got %d", pq.Len())
	}

	result := pq.Pop()
	if result == nil {
		t.Fatal("Pop() returned nil")
	}
	if result.ID != "task-1" {
		t.Errorf("Expected task-1, got %s", result.ID)
	}
	if pq.Len() != 0 {
		t.Errorf("Expected empty queue after Pop, got Len() = %d", pq.Len())
	}
}

func TestPriorityQueue_PopEmpty(t *testing.T) {
	pq := NewPriorityQueue()
	result := pq.Pop()
	if result != nil {
		t.Errorf("Expected nil from empty queue, got %v", result)
	}
}

func TestPriorityQueue_PeekEmpty(t *testing.T) {
	pq := NewPriorityQueue()
	result := pq.Peek()
	if result != nil {
		t.Errorf("Expected nil from empty queue, got %v", result)
	}
}

func TestPriorityQueue_PriorityOrder(t *testing.T) {
	pq := NewPriorityQueue()
	now := time.Now()

	tasks := []*Task{
		{ID: "low", Priority: PriorityLow, CreatedAt: now},
		{ID: "critical", Priority: PriorityCritical, CreatedAt: now},
		{ID: "medium", Priority: PriorityMedium, CreatedAt: now},
		{ID: "high", Priority: PriorityHigh, CreatedAt: now},
	}

	for _, task := range tasks {
		pq.Push(task)
	}

	expectedOrder := []string{"critical", "high", "medium", "low"}
	for i, expectedID := range expectedOrder {
		result := pq.Pop()
		if result == nil {
			t.Fatalf("Pop() returned nil at index %d", i)
		}
		if result.ID != expectedID {
			t.Errorf("Position %d: expected %s, got %s", i, expectedID, result.ID)
		}
	}
}

func TestPriorityQueue_SamePriority_FIFO(t *testing.T) {
	pq := NewPriorityQueue()

	// Create tasks with equal priority but different timestamps
	task1 := &Task{ID: "first", Priority: PriorityHigh, CreatedAt: time.Now()}
	task2 := &Task{ID: "second", Priority: PriorityHigh, CreatedAt: time.Now().Add(time.Millisecond)}
	task3 := &Task{ID: "third", Priority: PriorityHigh, CreatedAt: time.Now().Add(2 * time.Millisecond)}

	pq.Push(task3)
	pq.Push(task1)
	pq.Push(task2)

	// For equal priority, the oldest task should come first (FIFO)
	result1 := pq.Pop()
	result2 := pq.Pop()
	result3 := pq.Pop()

	if result1.ID != "first" {
		t.Errorf("Expected 'first', got '%s'", result1.ID)
	}
	if result2.ID != "second" {
		t.Errorf("Expected 'second', got '%s'", result2.ID)
	}
	if result3.ID != "third" {
		t.Errorf("Expected 'third', got '%s'", result3.ID)
	}
}

func TestPriorityQueue_Peek(t *testing.T) {
	pq := NewPriorityQueue()
	now := time.Now()

	pq.Push(&Task{ID: "low", Priority: PriorityLow, CreatedAt: now})
	pq.Push(&Task{ID: "high", Priority: PriorityHigh, CreatedAt: now})

	// Peek should return highest priority without removing
	result := pq.Peek()
	if result == nil {
		t.Fatal("Peek() returned nil")
	}
	if result.ID != "high" {
		t.Errorf("Peek() expected 'high', got '%s'", result.ID)
	}
	if pq.Len() != 2 {
		t.Errorf("Peek() should not remove items, Len() = %d", pq.Len())
	}
}

func TestPriorityQueue_Remove(t *testing.T) {
	pq := NewPriorityQueue()
	now := time.Now()

	pq.Push(&Task{ID: "task-1", Priority: PriorityLow, CreatedAt: now})
	pq.Push(&Task{ID: "task-2", Priority: PriorityMedium, CreatedAt: now})
	pq.Push(&Task{ID: "task-3", Priority: PriorityHigh, CreatedAt: now})

	// Entferne mittleren Task
	removed := pq.Remove("task-2")
	if !removed {
		t.Error("Remove() should return true for existing task")
	}
	if pq.Len() != 2 {
		t.Errorf("Expected Len() = 2 after remove, got %d", pq.Len())
	}

	// Entferne nicht-existierenden Task
	removed = pq.Remove("nonexistent")
	if removed {
		t.Error("Remove() should return false for non-existing task")
	}

	// Check remaining order
	first := pq.Pop()
	if first.ID != "task-3" {
		t.Errorf("Expected 'task-3', got '%s'", first.ID)
	}
	second := pq.Pop()
	if second.ID != "task-1" {
		t.Errorf("Expected 'task-1', got '%s'", second.ID)
	}
}

func TestPriorityQueue_Update(t *testing.T) {
	pq := NewPriorityQueue()
	now := time.Now()

	pq.Push(&Task{ID: "task-1", Priority: PriorityLow, CreatedAt: now})
	pq.Push(&Task{ID: "task-2", Priority: PriorityMedium, CreatedAt: now})

	// Update task-1 auf Critical
	pq.Update("task-1", PriorityCritical)

	// task-1 sollte jetzt zuerst kommen
	result := pq.Pop()
	if result.ID != "task-1" {
		t.Errorf("After update, expected 'task-1' first, got '%s'", result.ID)
	}
	if result.Priority != PriorityCritical {
		t.Errorf("Expected PriorityCritical, got %v", result.Priority)
	}
}

func TestPriorityQueue_ManyItems(t *testing.T) {
	pq := NewPriorityQueue()

	// Insert 100 tasks
	for i := 0; i < 100; i++ {
		priority := Priority(i % 4)
		pq.Push(&Task{
			ID:        fmt.Sprintf("task-%d", i),
			Priority:  priority,
			CreatedAt: time.Now().Add(time.Duration(i) * time.Millisecond),
		})
	}

	if pq.Len() != 100 {
		t.Errorf("Expected 100 items, got %d", pq.Len())
	}

	// Pop all and verify priority order is correct
	prevPriority := PriorityCritical
	for pq.Len() > 0 {
		task := pq.Pop()
		if task.Priority > prevPriority {
			t.Errorf("Priority order violated: got %v after %v", task.Priority, prevPriority)
		}
		prevPriority = task.Priority
	}
}
