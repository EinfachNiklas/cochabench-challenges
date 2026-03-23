package taskscheduler

import (
	"sync"
	"testing"
	"time"
)

func TestNewMetricsCollector(t *testing.T) {
	mc := NewMetricsCollector()
	if mc == nil {
		t.Fatal("NewMetricsCollector() returned nil")
	}

	snapshot := mc.Snapshot()
	if snapshot.TotalSubmitted != 0 {
		t.Errorf("Expected TotalSubmitted = 0, got %d", snapshot.TotalSubmitted)
	}
	if snapshot.TotalCompleted != 0 {
		t.Errorf("Expected TotalCompleted = 0, got %d", snapshot.TotalCompleted)
	}
}

func TestMetricsCollector_RecordSubmission(t *testing.T) {
	mc := NewMetricsCollector()

	mc.RecordSubmission(PriorityHigh)
	mc.RecordSubmission(PriorityHigh)
	mc.RecordSubmission(PriorityLow)

	snapshot := mc.Snapshot()
	if snapshot.TotalSubmitted != 3 {
		t.Errorf("Expected TotalSubmitted = 3, got %d", snapshot.TotalSubmitted)
	}
	if snapshot.TasksByPriority[PriorityHigh] != 2 {
		t.Errorf("Expected 2 high-priority submissions, got %d", snapshot.TasksByPriority[PriorityHigh])
	}
	if snapshot.TasksByPriority[PriorityLow] != 1 {
		t.Errorf("Expected 1 low-priority submission, got %d", snapshot.TasksByPriority[PriorityLow])
	}
}

func TestMetricsCollector_RecordCompletion(t *testing.T) {
	mc := NewMetricsCollector()

	mc.RecordCompletion(PriorityHigh, 100*time.Millisecond)
	mc.RecordCompletion(PriorityHigh, 200*time.Millisecond)

	snapshot := mc.Snapshot()
	if snapshot.TotalCompleted != 2 {
		t.Errorf("Expected TotalCompleted = 2, got %d", snapshot.TotalCompleted)
	}
	if snapshot.TasksByStatus[StatusCompleted] != 2 {
		t.Errorf("Expected 2 completed in status map, got %d", snapshot.TasksByStatus[StatusCompleted])
	}
}

func TestMetricsCollector_AverageExecTime(t *testing.T) {
	mc := NewMetricsCollector()

	mc.RecordCompletion(PriorityMedium, 100*time.Millisecond)
	mc.RecordCompletion(PriorityMedium, 300*time.Millisecond)

	snapshot := mc.Snapshot()

	// Gleitender Durchschnitt: (100 + 300) / 2 = 200ms
	expected := 200 * time.Millisecond
	diff := snapshot.AverageExecTime - expected
	if diff < 0 {
		diff = -diff
	}
	if diff > 10*time.Millisecond {
		t.Errorf("Expected AverageExecTime ~200ms, got %v", snapshot.AverageExecTime)
	}
}

func TestMetricsCollector_RecordFailure(t *testing.T) {
	mc := NewMetricsCollector()

	mc.RecordFailure(PriorityLow)
	mc.RecordFailure(PriorityHigh)

	snapshot := mc.Snapshot()
	if snapshot.TotalFailed != 2 {
		t.Errorf("Expected TotalFailed = 2, got %d", snapshot.TotalFailed)
	}
	if snapshot.TasksByStatus[StatusFailed] != 2 {
		t.Errorf("Expected 2 failed in status map, got %d", snapshot.TasksByStatus[StatusFailed])
	}
}

func TestMetricsCollector_RecordRetry(t *testing.T) {
	mc := NewMetricsCollector()

	mc.RecordRetry(PriorityMedium)
	mc.RecordRetry(PriorityMedium)
	mc.RecordRetry(PriorityMedium)

	snapshot := mc.Snapshot()
	if snapshot.TotalRetried != 3 {
		t.Errorf("Expected TotalRetried = 3, got %d", snapshot.TotalRetried)
	}
	if snapshot.TasksByStatus[StatusRetrying] != 3 {
		t.Errorf("Expected 3 retrying in status map, got %d", snapshot.TasksByStatus[StatusRetrying])
	}
}

func TestMetricsCollector_RecordStart(t *testing.T) {
	mc := NewMetricsCollector()

	mc.RecordSubmission(PriorityHigh) // Pending +1
	mc.RecordStart(PriorityHigh)      // Running +1, Pending -1

	snapshot := mc.Snapshot()
	if snapshot.TasksByStatus[StatusRunning] != 1 {
		t.Errorf("Expected 1 running, got %d", snapshot.TasksByStatus[StatusRunning])
	}
}

func TestMetricsCollector_FullLifecycle(t *testing.T) {
	mc := NewMetricsCollector()

	// Simuliere Lifecycle: Submit → Start → Complete
	mc.RecordSubmission(PriorityHigh)
	mc.RecordStart(PriorityHigh)
	mc.RecordCompletion(PriorityHigh, 50*time.Millisecond)

	// Simuliere: Submit → Start → Fail
	mc.RecordSubmission(PriorityLow)
	mc.RecordStart(PriorityLow)
	mc.RecordFailure(PriorityLow)

	snapshot := mc.Snapshot()
	if snapshot.TotalSubmitted != 2 {
		t.Errorf("Expected TotalSubmitted = 2, got %d", snapshot.TotalSubmitted)
	}
	if snapshot.TotalCompleted != 1 {
		t.Errorf("Expected TotalCompleted = 1, got %d", snapshot.TotalCompleted)
	}
	if snapshot.TotalFailed != 1 {
		t.Errorf("Expected TotalFailed = 1, got %d", snapshot.TotalFailed)
	}
}

func TestMetricsCollector_Snapshot_ReturnsCopy(t *testing.T) {
	mc := NewMetricsCollector()
	mc.RecordSubmission(PriorityHigh)

	snapshot1 := mc.Snapshot()

	// Modifiziere den Snapshot
	snapshot1.TotalSubmitted = 999
	snapshot1.TasksByPriority[PriorityHigh] = 999

	// Neuer Snapshot sollte unverändert sein
	snapshot2 := mc.Snapshot()
	if snapshot2.TotalSubmitted != 1 {
		t.Errorf("Snapshot should be a copy, but was modified: TotalSubmitted = %d", snapshot2.TotalSubmitted)
	}
	if snapshot2.TasksByPriority[PriorityHigh] != 1 {
		t.Errorf("Snapshot maps should be copies, but was modified: %d", snapshot2.TasksByPriority[PriorityHigh])
	}
}

func TestMetricsCollector_ConcurrentAccess(t *testing.T) {
	mc := NewMetricsCollector()
	var wg sync.WaitGroup

	// Concurrent Submissions
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			priority := Priority(id % 4)
			mc.RecordSubmission(priority)
		}(i)
	}

	// Concurrent Completions
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			mc.RecordCompletion(PriorityMedium, time.Duration(id)*time.Millisecond)
		}(i)
	}

	// Concurrent Snapshots
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mc.Snapshot()
		}()
	}

	wg.Wait()

	snapshot := mc.Snapshot()
	if snapshot.TotalSubmitted != 100 {
		t.Errorf("Expected 100 submissions, got %d", snapshot.TotalSubmitted)
	}
	if snapshot.TotalCompleted != 50 {
		t.Errorf("Expected 50 completions, got %d", snapshot.TotalCompleted)
	}
}
